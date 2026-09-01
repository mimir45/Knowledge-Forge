# Knowledge Forge — Mimari

> Bu doküman **nasıl** sorusunu cevaplar. *Neden* için [`01-FIKIR.md`](01-FIKIR.md),
> kullanım için [`03-KULLANIM-KILAVUZU.md`](03-KULLANIM-KILAVUZU.md), dosya dökümü için
> [`04-DOSYA-DOSYA.md`](04-DOSYA-DOSYA.md).

---

## 1. Kuşbakışı

Sistem dört katmandan oluşur. Alttan üste doğru okunmalı, çünkü üst katmanların hiçbiri
alttakiler olmadan anlamlı değil, tersi doğru değil.

```
┌──────────────────────────────────────────────────────────────────────┐
│  KATMAN 3 — Claude Code entegrasyonu                                 │
│  .claude-plugin/  hooks/  skills/  agents/                           │
│  (lifecycle hook'ları, slash command'lar, product agent spec'leri)   │
└──────────────────────────────────────────────────────────────────────┘
                              │  exec
┌──────────────────────────────────────────────────────────────────────┐
│  KATMAN 2 — CLI:  cmd/forge  (20 subcommand)                         │
│  flag parse · orkestrasyon · çıktı biçimlendirme · exit code         │
└──────────────────────────────────────────────────────────────────────┘
                              │  import
┌──────────────────────────────────────────────────────────────────────┐
│  KATMAN 1 — Kütüphaneler:  pkg/*  (18 paket)                         │
│  tüm iş mantığı; her biri tek başına test edilebilir                 │
└──────────────────────────────────────────────────────────────────────┘
                              │  read/write
┌──────────────────────────────────────────────────────────────────────┐
│  KATMAN 0 — Veri                                                     │
│  vault/*.md (doğruluk kaynağı) · kod repoları (git) ·                │
│  .forge/*.db + .forge/*.json (türetilmiş cache)                      │
└──────────────────────────────────────────────────────────────────────┘
```

**Öğretici not:** Bu ayrımın gerçek testi şudur — `cmd/forge` içindeki hiçbir dosya iş
mantığı içermez. `cmd/forge/drift.go` 250 satırdır ama tek bir verdict kararı vermez;
flag'leri okur, `pkg/drift`'i çağırır, sonucu JSON ya da metin olarak basar. Bu yüzden
`pkg/drift`'in testleri bir CLI olmadan çalışır ve CLI'ın testleri iş mantığını yeniden
test etmez.

---

## 2. Paket haritası ve import DAG'ı

18 paket var. Aralarındaki bağımlılıklar **döngüsüz** ve şaşırtıcı derecede seyrek:

```
cmd/forge ──────────────► (18 pkg'nin hepsi) + profiles

pkg/config    ──► config (embed)
pkg/vault     ──► pkg/config, references
pkg/dataset   ──► pkg/vault
pkg/scrub     ──► pkg/vault
pkg/engine    ──► pkg/config
pkg/drift     ──► pkg/codeindex, pkg/coderef, pkg/vault
pkg/qualitygate ──► pkg/config, pkg/recall, pkg/similarity, pkg/vault, references
pkg/report    ──► pkg/drift, pkg/gitsig, pkg/graph, pkg/linkcheck, pkg/similarity

yapraklar (hiçbir iç paketi import etmez):
  codeindex · coderef · gitsig · graph · linkcheck · recall
  sentinel  · similarity · store · telemetry
```

Bu grafikten okunacak üç şey var:

**(a) On paket yapraktır.** `pkg/recall` hiçbir şey import etmez — ne vault, ne store.
Girdisi `[]recall.Doc`, çıktısı `recall.Result`. Bu, skorlama algoritmasını saf bir
fonksiyon haline getirir ve `pkg/recall/score_test.go`'nun neden bir vault kurmadan
çalışabildiğini açıklar.

**(b) `pkg/report`, `pkg/codeindex`'i import ETMEZ** — bu yazılı bir kuraldır, kaza
değil. Nedeni cgo: `pkg/codeindex` tek cgo paketidir (go-tree-sitter). Rapor katmanı onu
import etseydi, saf-Go build lane'i çökerdi. Bunun yerine `cmd/forge/check_codebase.go`
bir `symbolFinder` arayüzü tanımlar ve somut tipi CLI'da birleştirir.

**(c) `pkg/drift` cgo'lu `pkg/codeindex`'i import ediyor gibi görünse de**, kendi
`Source` arayüzü sayesinde saf Go kalır:

```go
type Source interface {
    At(...)        // bir commit'teki dosya içeriği
    RevBefore(...) // önceki revizyon
    Head(...)      // HEAD sha
    Find(...)      // sembol arama
    ResolveAt(...) // belirli bir era'da çözümleme
}
```

Arayüz tam olarak bunun için var: drift'in mantığı saf Go'da test edilebilir, cgo'yu
taşıyan somut implementasyon dışarıda kalır.

### 2.1 Paketlerin tek cümlelik görevleri

| Paket | Görev |
|---|---|
| `pkg/vault` | Frontmatter + markdown AST (goldmark), mtime-cache'li. Not okuma/yazma/doğrulama/karantina. |
| `pkg/recall` | Deterministik soru → not skorlaması. Sıfır model çağrısı. |
| `pkg/similarity` | Elle yazılmış MinHash + LSH banding. Yakın-duplicate tespiti. |
| `pkg/graph` | Not link grafiği: bileşenler, hub'lar, orphan'lar, merkezilik. |
| `pkg/codeindex` | go-tree-sitter (Java + TypeScript). **Tek cgo paketi**, build-tag'li. |
| `pkg/coderef` | Not gövdesinden ve frontmatter'dan kod atıflarını çıkarır ve çözer. |
| `pkg/gitsig` | Churn, ownership, co-change coupling — git CLI üzerinden (go-git değil). |
| `pkg/drift` | **Anahtar paket.** AST karşılaştırması ile çürüme tespiti, satır diff'i değil. |
| `pkg/linkcheck` | Kaynaklara HTTP HEAD, cache'li, rate-limit'li. |
| `pkg/report` | Analizleri markdown'a render eder. `pkg/codeindex`'i import etmemeli. |
| `pkg/store` | SQLite (`modernc.org/sqlite`). Budget tablosu hariç sadece türetilmiş cache. |
| `pkg/config` | Dört katmanlı config zinciri. |
| `pkg/engine` | none/host/api/advisor backend'leri, stage başına seçim + fallback, engine_trail. |
| `pkg/qualitygate` | Yedi DESIGN §12 kapısı + `Run`/`Report` orkestrasyonu + `_inbox/` karantinası. |
| `pkg/sentinel` | Id'li begin/end yönetilen yorum blokları; `Upsert`/`UpsertBefore`/`Remove`. |
| `pkg/scrub` | Bir vault kopyasından sır/PII biçimli içeriği redakte eder; kapalı devre hata verir. |
| `pkg/dataset` | D2–D4 eğitim çifti yakalama (JSONL). |
| `pkg/telemetry` | DESIGN §14'ün `ask` olayı; sha256 konu hash'i, asla ham soru metni. |

---

## 3. İki build lane'i

Bu, mimarinin en sık yanlış anlaşılan yeridir. **İki ayrı derleme modu var**, ve
`pkg/codeindex` ikisini ayıran şey.

```
LANE A — saf Go (varsayılan, dağıtılan)
  CGO_ENABLED=0 go build ./...
  → tree-sitter YOK, kod indeksleme devre dışı
  → altı hedefe cross-compile eder
  → make build

LANE B — cgo
  CGO_ENABLED=1 go build -tags codeindex ./...
  → go-tree-sitter derlenir, Java + TypeScript parse edilir
  → host toolchain gerekir, cross-compile etmez
  → make full
```

Mekanizma iki dosyalı klasik Go deseni:

- `pkg/codeindex/parse_cgo.go` — gerçek tree-sitter parser'ı, build tag'li.
- `pkg/codeindex/parse_nocgo.go` — aynı imzalar, boş/degrade davranış.

Böylece `pkg/codeindex`'i import eden kod her iki lane'de de **derlenir**; sadece
davranış farklıdır. İlgili invariant:

> `CGO_ENABLED=0` her paket için, `pkg/codeindex` hariç (go-tree-sitter cgo gerektirir).

`make test` bunu bilerek her iki yönden zorlar: önce `CGO_ENABLED=1 go test ./...`
(gerçek parser'la), sonra `CGO_ENABLED=0 go build ./...` (saf lane'in hâlâ derlendiğini
kanıtlamak için). Sadece birini çalıştırmak, diğerini sessizce kırar.

Ölçülmüş durum: 18 paket `ok` raporluyor, her iki `CGO_ENABLED` değerinde de yeşil.
(`config`, `profiles`, `references` yalnızca veri paketleri — test dosyaları yok.)

---

## 4. Config: dört katmanlı zincir

```
1. $FORGE_CONFIG                          (en yüksek öncelik)
2. <project>/.forge.config.md
3. ~/.forge/forge.config.md
4. config/forge.config.example.md         (binary'ye gömülü, en düşük)
```

Kurallar, `pkg/config/load.go` ve `merge.go`'da:

- **Eksik bir opsiyonel katman atlanır.** Ama **eksik bir `$FORGE_CONFIG` hatadır** —
  çünkü kullanıcı onu açıkça isimlendirmiştir. Sessizce yok saymak, yanlış config ile
  çalıştığını fark etmemeye yol açar.
- **Map'ler anahtar anahtar birleşir.** Scalar'lar ve list'ler bütünüyle değiştirilir.
  Yani `engines.budget.api_usd_per_day`'i override etmek için tüm `engines` bloğunu
  yeniden yazmanız gerekmez; ama `check.reports` listesinin bir elemanını eklemek için
  listenin tamamını yazmanız gerekir.
- Config bir **markdown dosyasıdır**, frontmatter'ı YAML. `frontmatter()` BOM'u temizler
  ve CRLF'i normalize eder — Windows'tan gelen bir config sessizce parse hatası vermesin
  diye.
- `decode()` YAML üzerinden round-trip yapar, *"böylece struct tag'leri şemanın tek
  tanımı olarak kalır"* — yani şema iki yerde tanımlanmaz.
- `expandHome()` `~/` çözer.

**Yazıcı kim?** Sadece `forge init`. Ve sadece iki dosyaya:
`~/.forge/forge.config.md` ve `<vault>/profiles/me.md`. **Asla**
`config/forge.config.md`'ye — o paketlenmiş bir şablon olarak kalır. Bu ayrımın ihlali,
bir sonraki `go build`'in kullanıcının ayarlarını ezmesi demek olurdu.

Şema, ADDENDUM §E ile DESIGN §10'un **birleşimidir** — §E'nin yeniden ifade etmediği
§10 anahtarları da geçerlidir.

`forge config --layers` hangi katmanların bulunduğunu, `forge config --json` çözülmüş
sonucu basar. Bir "neden bu değer?" sorusunda ilk çalıştırılacak komut budur.

### 4.1 Config ağacının ana blokları

`config/forge.config.example.md` (215 satır) tam varsayılan ağacı taşır:

| Blok | İçerik |
|---|---|
| `vault_path`, `repo_path: auto`, `paths` | Nerede çalışılacağı. |
| `trigger.mode: ask` | "explain X" anında sorulsun mu, otomatik mi. |
| `recall` | Leksikal; eşikler **0.85 / 0.55**, duplicate 0.30. |
| `freshness_days` | Tipe göre: concept 365, howto 180, api 90, pattern 365, pitfall 365, incident 0, decision 0. |
| `engines` | default `host`; api `anthropic`/`claude-sonnet-5`/`ANTHROPIC_API_KEY`; advisor `claude-opus-5` mode `critique`; local kapalı; bütçe 2.00/1.00 USD; `on_exhausted: queue`; `routing.advisor_when`. |
| `pipeline` | Dokuz stage (aşağıda). |
| `research`, `verify`, `write` | Araştırma derinliği; `run_code: auto`, `duplicate_threshold: 0.40`; dil `en`, 1200 kelime, mermaid. |
| `static` | code_index (java/kotlin/python/typescript), git_signals, `cache_ttl_days: 30`, drift, linkcheck, logback (`inline_markers: false`). |
| `check` | `schedule: "0 9 * * MON"`, dokuz rapor, churn 90 gün, duplicate 0.40. |
| `garden`, `dataset`, `telemetry` | Bahçıvanlık; D1–D5 yakalama + `anonymize_on_export`; telemetri `local` kapsam. |

Dokümanın prose bölümü ayrıca **koda bilerek bırakılan** üç grubu kaydeder:
`pkg/vault`'un `excludedPrefixes`/`hubNames`'i, `pkg/report/duplicates.go`'nun
`specThreshold = 0.85`'i, ve Makefile'daki `$HOME/.forge/bin`. Bunlar config'e
çıkarılmadı çünkü bir kullanıcı kararı değil, bir spesifikasyon sabiti.

---

## 5. Engine katmanı: tier'lar, zincirler ve bütçe

### 5.1 Dört tier

| Tier | Implementasyon | Not |
|---|---|---|
| `none` | `pkg/engine/none.go` | Model yok. |
| `host` | `host.go` | Oturumdaki Claude Code. |
| `api` | `api.go` + `api_provider.go` | Anthropic API. |
| `advisor` | `advisor.go` | Kritik-only. |
| `local` | *alias* | **Beşinci engine değil** — "farklı bir `base_url` altında `api.go`". `TierAPI`'ye map'lenir. |

### 5.2 Seçim algoritması

`pkg/engine/select.go`'daki `Resolve()`, stage'in zincirini yürür:

```go
func chain(cfg, stage, st) []string {
    if st.Engine != "" { out = append(out, st.Engine) }
    else if cfg.Engines.Default != "" { out = append(out, cfg.Engines.Default) }
    if st.Fallback != "" { out = append(out, st.Fallback) }
    if st.Then != "" { out = append(out, st.Then) }
    if len(out) == 0 { out = []string{"none"} }
    return out
}
```

Öğretici detay, kaynağın kendi yorumunda: *"ayarlanmamış bir stage, none'a kilitlenmiş
olma iddiası değildir; sessizliktir — ve onu dolduran `cfg.Engines.Default`'tur."*
Boş bir alanı "none" saymak ile "varsayılan" saymak arasındaki fark, config'i yazan
kişinin niyetine dair bir yorumdur ve burada açıkça yapılmıştır.

`Resolve` sadece kazanan ismi değil, **insan-okunur bir gerekçe** de döndürür. Bu,
`forge engine select --json`'un "offline neden none'a düştü" sorusunu cevaplayabilmesini
sağlar — sadece "düştü" demesini değil. Hiçbir aday uygun değilse:
`"no candidate in the chain was available; degrading to none"`.

### 5.3 Kilitli stage'ler — iki katmanlı savunma

Üç stage `none` dışında bir şey kabul etmez: **`recall`, `write`, `index`**.

Kontrol iki yerde:

1. **Yükleme anında** — `pkg/config/validate.go`, `LockedStageError`.
2. **Seçim anında** — `pkg/engine/select.go`, `checkLocked()`.

İkincisi gereksiz görünebilir ama kaynak gerekçesini yazıyor: `checkLocked`, `Engine`'in
yanı sıra **`Fallback` ve `Then`**'e de bakar — *"`pipeline.write.engine` yerine
`pipeline.write.fallback`'in arkasına saklanan bir tamper burada da yakalanmalı, yoksa bu
katman dekoratiftir."*

Hata mesajı da eğiticidir, sessiz bir override değil:

```
engine: pipeline.write: "api" is not allowed — [recall write index] are locked to
"none" (T0 static core)
```

### 5.4 Bütçe muhasebesi

- Sayaçlar **SQLite'ta** (`pkg/store/budget.go`), `.forge/` altında.
- `forge reindex`'ten **sağ çıkar** — cache'in tek istisnası budur. Aksi halde reindex,
  bütçe sıfırlama hilesi olurdu.
- `Exhausted()` "bugün bütçe yok" ile "burada hiç ölçülen tier yok" arasını ayırır;
  `on_exhausted: queue` bu ayrım olmadan yanlış çalışırdı.
- `on_exhausted` varsayılanı **`queue`**. Kabul edilen üç değer `queue | degrade | stop`:
  `queue` bir sonraki bütçe döngüsünde işlenmek üzere `pending_advisor: true` damgalar ve
  `none` tier'a düşer; `degrade` bugün `none`'a sessizce düşmekle aynı şey, çünkü kelimenin
  dürüst okuması zaten bu; `stop` gerçek, sıfır olmayan bir çıkışla durur — `pkg/engine`'in
  kendisi `OnExhausted`'ı okumaz, bu ayrım bir katman yukarıda `cmd/forge`'da yapılır.

### 5.5 engine_trail

`pkg/engine/trail.go`, hangi stage'in hangi tier'a gittiğini notun frontmatter'ında
kaydeder. `forge engine record` bunu yazar ve **kilitli stage'e kayıt yapmayı
reddeder** (`isLockedStage`). Böylece "bu not nasıl üretildi?" sorusu notun kendisinden
cevaplanır.

---

## 6. Recall motorunun anatomisi

`pkg/recall/score.go` (246 satır) sistemin en yoğun düşünülmüş dosyasıdır.

### 6.1 Dört kanal

```go
const (wTitle = 0.4; wTags = 0.3; wStack = 0.2; wBody = 0.1)
```

### 6.2 Karıştırma (blend) — aktif kanallar üzerinden ortalama

```go
func blend(chs []Channel) (score float64, matched []string) {
    num, den := 0.0, 0.0
    for _, c := range chs {
        if !c.Active { continue }
        num += c.Weight * c.Value
        den += c.Weight
        if c.Value > 0 { matched = append(matched, c.Name) }
    }
    if den == 0 { return 0, matched }
    return num / den, matched
}
```

`Active == false` olan kanal hem paydan hem paydadan düşer. `weighted()` boş payda
durumunda `ok=false` döndürür ve kanal devre dışı kalır.

### 6.3 IDF

```go
const idfCap = 3.5

func idf(df, n int) float64 {
    if df <= 0 || n <= 0 { return 0 }
    return math.Min(math.Log(1+float64(n)/float64(df)), idfCap)
}
```

`log(1+n/df)`, `log(n/df)` değil — ikincisi bir terim her notta geçtiğinde tam sıfır
verir ve terimi denklemden tamamen siler.

### 6.4 Başlık: F₂

```go
func f2(hits, queryTerms, titleTokens int) float64 {
    if hits == 0 || queryTerms == 0 || titleTokens == 0 { return 0 }
    p := float64(hits) / float64(titleTokens)
    r := float64(hits) / float64(queryTerms)
    return 5 * p * r / (4*p + r)
}
```

### 6.5 Gövde

`bodyChannel` her terimi **3 tekrarda doyurur** (saturate). Bir kelimeyi 50 kez geçen
bir not, 3 kez geçenden daha alakalı değildir.

### 6.6 IDF ağırlıklandırmasının bulduğu ve düzelttiği kalibrasyon açığı

İlk IDF ağırlıklandırması shipped edildiğinde hedeflediği vakayı düzeltmemişti. Neden:
bir sorunun anlamını taşıyan terimler, hiçbir not onları taşımadığında paydadan filtrelenir
— yani "kimse bilmiyor" durumu, "önemsiz" gibi davranıyordu. Düzeltme iki değişiklik
istedi: `inVocab` filtresi soruya değil `--stack` ipucuna uygulanacak şekilde taraf
değiştirdi, ve eksik bir terimin ağırlığı sıfır yerine mevcut terimlerin ortalaması oldu.
**Eşikleri buna cevaben oynatmayın** — düzeltme §3.1 kalibrasyon tablosunun ölçülerek
yeniden türetilmesiyle geldi, sabit değiştirerek değil.

---

## 7. Drift: git-anchored çürüme tespiti

### 7.1 Sözleşme

```go
type Verdict string
const (OK; Repaired; Suspect; Broken; Skipped)

func (f Finding) Demoting() bool { return f.Verdict == Broken }

type Changed struct {
    Touched map[string]bool
    Deleted map[string]string  // repo-relative path -> repo adı
}

type Opts struct{ Deep bool }
```

### 7.2 Neden git object store, working tree değil

Drift `HEAD` ağacını ve `--since-commit <sha>`'yı okur. Çalışma ağacına **bakmaz**.
Nedeni:

- Yarı yazılmış bir dosya bir iddia değildir. Bir not, henüz kaydedilmemiş bir
  düzenleme yüzünden demote edilmemelidir.
- Determinizm: (not atıfları, ağaç durumu) → verdict saf bir fonksiyondur. Working tree
  bu fonksiyonu zamana bağımlı hale getirirdi.
- Simetri: `git revert` → aynı ağaç → aynı verdict → not geri yükselir.

`.forge/`'da saklanan tek state, demote'tan önceki confidence değeri — bir geri yükleme
hedefi, asla bir verdict girdisi.

### 7.3 İki yol: hook yolu ve full sweep

| | Hook yolu (`forge drift`, varsayılan) | Full sweep (`forge check`) |
|---|---|---|
| Tetikleyici | post-commit / post-merge / post-checkout | Haftalık |
| Kapsam | `--since-commit` ile değişen dosyalar | Tüm vault |
| Registry | `HEAD` ağacından | `HEAD` + (`--deep` ile) tarihsel `ResolveAt` |
| `drift.Apply` çağırır mı? | `--apply` ile evet | **Hayır** |
| Bütçe | **< 100 ms** (bağlayıcı kısıt) | < 10 s |

Bu ayrım silinmiş-dosya atıflarının nasıl `Broken` verdict aldığını anlamak için önemli:

`registryOf` registry'yi her zaman güncel `HEAD` ağacından kurduğu için, tamamen silinmiş
bir dosyaya yapılan atıf asla `Broken` veremiyordu — sonsuza kadar `Skipped` kalıyordu.
Çözüm iki parçalı: full sweep'te (`opts.Deep`, `--since-commit` yok) doğrulanmış-era bir
`ResolveAt` taramasına düşmek — ama `forge check`'in full sweep'i `drift.Apply`'ı hiç
çağırmadığı için bu tek başına `drift.md`'yi doğru yapar, **hiçbir şeyi otomatik demote
etmez.** Otomatik demote hook yolunda oluyor: hook zaten ucuz bir kapı hesaplıyordu
(`coderef.ChangedFilesStatus`); `--name-only` yerine `--name-status` kullanılarak bu kapı
**silme kanıtı** taşır hale getirildi (`drift.Changed{Touched, Deleted}`). Artık aynı
commit'teki bir silmeyle eşleşen `Unresolved` bir atıf, `--apply` altında **anında**
`Broken` verdict alır — `--deep` ve tarihsel registry taraması gerekmeden. Bu, silinmiş-
dosya atıfının sahip olduğu tek otomatik demote yoludur.

Mimari açıdan kritik bir ayrıntı: hook yolunda eşleşmeyen bir kaçırma **hiç finding
üretmez**, asla `Skipped` üretmez. Aksi halde alakasız, sonraki bir commit hâlâ bozuk bir
notu `high`'a geri çevirebilirdi. `TestRollbackSymmetryOnDeletion` bunu sabitler.

Bu yaklaşımın bilinen bir sınırı var: basename çakışması.

### 7.4 AST karşılaştırması, satır diff'i değil

Bir dosyanın 200 satırı değişmiş olabilir ama atıf yapılan sembol hiç dokunulmamış
olabilir; ya da tek satır değişmiştir ve o satır metot imzasıdır. Satır diff'i ikisini
ayırt edemez, AST karşılaştırması eder. `pkg/codeindex`'in tree-sitter'ı burada devreye
girer — ve tam da bu yüzden cgo gerektirir.

---

## 8. Kalite kapıları ve karantina

`pkg/qualitygate/gate.go`'nun `Run` sırası sabittir:

```
schema → citation → code → freshness → antislop → link → duplicate
```

`Remedy` bir iota'dır ama JSON'a **ordinal değil isim** olarak serialize edilir
(`MarshalJSON`) — bir sonraki sürümde araya bir remedy eklemek, kaydedilmiş raporları
bozmasın diye.

```go
func blocksWrite(r Remedy) bool  // None, DelegateToLibrarian, SwitchToUpdate → false
```

Blocking bir hata → `_inbox/` karantinası, `confidence: low`. `cmd/forge/gate.go`
karantina sonrası `reindexAfterQuarantine` çağırır: index, notun artık `_inbox/`'ta
olduğunu yansıtmalı.

`code` kapısı ilginç bir alt sisteme sahip: `pkg/qualitygate/compile*.go` —
`compile_bash.go`, `compile_java.go`, `compile_ts.go`. Bunlar sistem toolchain'ini
çağırır, **tek kullanımlık bir dizinde**. Ölçülen maliyet, kapı mantığı değil toolchain
başlangıcıdır: bash ~10 ms sıcak (~470 ms soğuk, tek seferlik OS page-cache etkisi),
java ~170 ms sıcak (~370 ms soğuk). `tsc` bu ortamda kurulu olmadığı için TypeScript
şeridi test edilemedi; `TestCompileTSSkippedWhenToolchainAbsent` bunun yerine
toolchain-yok yolunu kapsıyor.

`code` hariç altı in-process kapı toplam **~0.13 ms** sürüyor — yani kalite kontrolü
pratik olarak bedava, tek maliyet derleyici çağırmak.

---

## 9. Sentinel: idempotent yönetilen bloklar

30 satırlık bir paket, ama `logback`'in tamamı buna dayanıyor.

```go
type Style struct{ Open, Close string }

var (
    Markdown = Style{"<!--", "-->"}
    Slash    = Style{"//"}
    Hash     = Style{"#"}
)
```

İşaretler `forge:<id>:begin` / `forge:<id>:end` olarak render edilir. Sözleşme tek
cümlede: *"Bir bloğun kendi begin/end çiftinin dışındaki her şey, bayt bayt dokunulmadan
bırakılır."*

`Upsert` / `UpsertBefore` / `Remove` — idempotent ve **konumdan bağımsız.** Yani
kullanıcı bloğu dosyada başka bir yere taşırsa, bir sonraki `logback` çalışması onu
oraya yazmaya devam eder, ikinci bir kopya oluşturmaz.

Bu, "üretilen içeriği kullanıcının dosyasına yazmak" probleminin doğru çözümüdür:
tüm dosyayı sahiplenmek yerine, dosyanın adı verilmiş bir bölümünü sahiplenmek.

---

## 10. Üç ana veri akışı

### Akış A — Soru → Not

```
kullanıcı: "explain X"
   │
   ├─ [hook] UserPromptSubmit → forge intent
   │     stdin'den prompt oku → recall → skor > 0.7 ise
   │     additionalContext olarak en iyi hit'i bas
   │     bütçe < 50 ms · fail-silent · exit 0
   │
   ▼
forge recall --explain
   │  pkg/vault ile notları yükle (SQLite cache'ten sıcak)
   │  pkg/recall ile skorla (4 kanal, IDF, F₂)
   │  verdict: reuse | update | create
   │  pkg/telemetry: ask olayı (sadece konu hash'i)
   ▼
verdict = create
   │
   ├─ pipeline: intake → plan → research → synthesize → verify → write
   │  (her stage kendi engine tier'ını pkg/engine'den seçer)
   ▼
taslak not
   │
   ▼
forge gate --file draft.md
   │  yedi kapı, sırayla
   │  blocking hata? → _inbox/, confidence: low → reindex
   │  temiz? → notes/<type>/<slug>.md
   ▼
forge index → _index.md + SQLite
```

### Akış B — Commit → Drift

```
kod reposunda git commit
   │
   ▼
.git/hooks/post-commit  (scripts/install_drift_hook.sh ile kurulur)
   │
   ▼
forge drift --repo <name>:<path> --since-commit <sha> --apply
   │
   ├─ coderef.ChangedFilesStatus (--name-status)
   │     → drift.Changed{Touched, Deleted}          ← ucuz kapı
   │
   ├─ hiçbir atıf etkilenmiyorsa: çık (yaygın durum, ~60 ms)
   │
   ├─ etkilenen atıflar için:
   │     registryOf(HEAD ağacı) → coderef çözümleme
   │     AST karşılaştırması (pkg/codeindex, tree-sitter)
   │     → OK | Repaired | Suspect | Broken
   │
   ▼
Broken → demote
   │  önceki confidence .forge/'a kaydedilir (geri yükleme hedefi)
   │  notun confidence'ı düşürülür
   │
   ▼
git revert → aynı ağaç → aynı verdict → not simetrik olarak geri yükselir
```

Bütçe **< 100 ms** ve ölçülen **60–70 ms**. Bu, tüm projenin bağlayıcı latency
kısıtıdır — çünkü git hook yolunda çalışan tek şey budur.

### Akış C — Haftalık kontrol

```
forge check   (schedule: "0 9 * * MON", ama zamanlanmış otomatik mutasyon YOK)
   │
   ├─ collectVault: notlar, graph, similarity, git history, budget snapshot
   │
   ├─ dokuz raporu render et → <vault>/reports/
   │     coverage · staleness · duplicates · orphans · gaps
   │     graph-health · churn · deadlinks · drift
   │     (+ cost.md, codebase.md, moc/weekly/YYYY-WW.md)
   │
   ├─ weekly rollup: .forge/weekly-stats.json ile hafta-üstü-hafta delta
   │
   ├─ aiPass (opsiyonel): draft refresh · duplicate merge · ADR stub önerileri
   │     — sadece TALİMAT basar, kendisi yazmaz
   │
   └─ drainAdvisorQueue: bütçe döndüyse kuyruktaki notları işle
```

Ölçülen: **390 ms sıcak / 930 ms soğuk**, bütçe 10 s. Dokuz rapor deterministik render
ediliyor — altı ardışık çalıştırma, md5-özdeş.

Gerçek vault'a karşı ölçülen sonuçlar: **9 not** değişmiş koda atıf yapıyor (2 broken,
7 suspect), 140 atıf üzerinden; **94 notun 21'i** orphan; **23 graph bileşeni**;
**3 duplicate çifti** ≥ 0.40; **41 stack'in 39'u** kapsanmış.

`writeReport` sadece içerik değiştiyse yazar (`writeIfChanged`) — böylece değişmemiş bir
raporun mtime'ı oynamaz ve vault'un git diff'i temiz kalır.

---

## 11. Claude Code entegrasyon katmanı

### 11.1 Plugin manifest

`.claude-plugin/plugin.json` — name `forge`, displayName "Knowledge Forge", v0.1.0, MIT,
repo `github.com/mimir45/Knowledge-Forge`.
`.claude-plugin/marketplace.json` — marketplace kaydı.

Bu ikisi Phase 6'ya kadar var olmayan **paketleme boşluğunu** kapatır: kök seviyedeki
`agents/` dizini ve `hooks/hooks.json`, plugin kurulduğunda otomatik keşfedilir.

### 11.2 Dört lifecycle hook'u

| Event | Matcher | Komut | Timeout |
|---|---|---|---|
| `SessionStart` | — | `hooks/session-context` | 5 s |
| `UserPromptSubmit` | — | `hooks/user-prompt-intent` | 2 s |
| `SessionEnd` | — | `hooks/session-end-capture` | 10 s |
| `PostToolUse` | `WebFetch` | `hooks/post-tool-cache-source` | 10 s |

Yollar `"${CLAUDE_PLUGIN_ROOT}"/hooks/...` — Phase 6'da hardcode edilmiş mutlak
yollardan buna geçildi.

**Ortak sözleşme:** hepsi fail-silent, **her zaman exit 0.** Bir hook, oturumu ya da
commit'i asla kıramaz.

**Resume tuzağı** (kaynakta belgeli): `SessionStart` `--continue`/`--resume`'da yeniden
çalışır (`source: resume` ile) — beklenen, çıktısı idempotent ve ucuz. Ama **diğer her
hook'un çıktısı kaydedilmiş transcript'ten tekrar oynatılır**, yeniden çalıştırılmaz.
Sonuç: resume'da bayat bir recall hit'i görmek `forge intent`'te bir bug değil, beklenen
davranıştır. Bu yüzden hiçbir hook'a zamana duyarlı iş konmaz.

### 11.3 Git hook'ları

Üç ayrı hook ailesi var, karıştırılmamalı:

| Hook | Nerede | Ne yapar | Kurulum |
|---|---|---|---|
| `vault-post-commit` | **vault** reposunda | `forge capture` — D3 eğitim çifti hasadı | `scripts/install_vault_hook.sh` |
| `code-post-commit` / `-merge` / `-checkout` | **kod** repolarında | `forge drift` | `scripts/install_drift_hook.sh` |
| `hooks/hooks.json`'daki dört | Claude Code oturumunda | context/intent/capture/cache | plugin kurulumu |

Vault hook'u **`~/.forge/bin/forge`**'u çağırır — repo'nun build çıktısını değil. Mutlak
yol `<vault>/.forge/forge-bin`'de pinlenmiştir, `$FORGE_BIN` override eder. Bu bir
**kopyadır**: `pkg/dataset` ya da `cmd/forge/capture.go` değiştiğinde yeniden kurmak
gerekir:

```bash
CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge
```

Hook tasarım gereği hiçbir şey basmadığı için bayat bir binary **sessizdir**. Teşhis:
`<vault>/.forge/capture.log`.

### 11.4 Dört skill

`skills/forge/`, `skills/forge-init/`, `skills/forge-check/`, `skills/forge-stats/`.
Skill'ler soruları sorar ve binary'ye shell out eder — iş mantığı taşımazlar.

### 11.5 Dört product agent

`agents/forge-researcher.md`, `forge-codebase-scout.md`, `forge-verifier.md`,
`forge-librarian.md`.

`forge-librarian`'ın prompt'u, authored ettiği her commit'e **`Forge-Write: true`**
damgası basar. Bu kritik: aksi halde `pkg/dataset` agent'ın kendi çıktısını
*insan düzeltmesi* olarak kaydeder ve eğitim verisi kirlenir.
`pkg/dataset/d3_forge_write_test.go` guard'ı iki yönlü sabitler.

**Kaydedilmiş boşluk:** bunlar `.claude/agents/`'taki **workflow** agent'larıyla
karıştırılmamalı. `.claude/agents/` altındakiler (`finder`, `executor`, `explainer`,
`vault-analyst`, `doc-auditor`, `cross-checker`) bu projeyi *inşa etmek* içindir.
`agents/` altındakiler *ürünün* agent'larıdır.

---

## 12. Invariant tablosu

Her satır farklı bir dokümanda ifade edilmiş ve her biri kazayla ihlal edilmesi kolay.

| # | Invariant | Nerede zorlanır |
|---|---|---|
| 1 | T0 statik çekirdek sıfır model çağrısı yapar. | `cmd/forge/main.go` doc comment; kod incelemesi |
| 2 | `recall`/`write`/`index` sadece `none`; aksi halde **net hatayla başlamayı reddet**. | `pkg/config/validate.go` + `pkg/engine/select.go:checkLocked` |
| 3 | Drift git-anchored; asla dosya kaydında, asla working tree'ye karşı. Verdict'ler saf fonksiyon; revert simetrik geri yükler. Demote geçmişi `.forge/`'da, asla not gövdesinde. | `pkg/drift`, `rollback_test.go` |
| 4 | `CGO_ENABLED=0` her paket için, `pkg/codeindex` hariç. | `parse_cgo.go`/`parse_nocgo.go`, `make test` |
| 5 | Markdown tek doğruluk kaynağı; SQLite türetilmiş; `forge reindex` tamamen yeniden kurar. | `cmd/forge/index.go:cmdReindex` |
| 6 | `pkg/similarity` elle yazılmış MinHash + LSH. **Embedding yok.** | `pkg/similarity/*` |
| 7 | Vault'u asla zamanlamayla otomatik mutasyona uğratma; kapı hataları `_inbox/`'a `confidence: low`. | `pkg/qualitygate/quarantine.go` |
| 8 | Kod doğrulama tek kullanımlık dizinde derler, asla kullanıcının projesinde. | `pkg/qualitygate/compile.go` |
| 9 | Advisor tier'ı kritik-only: tartışmalı iddialar + patch, asla yeniden yazım. | `pkg/engine/advisor.go` |
| 10 | Telemetri konu + hash loglar. Asla ham soru, kod ya da dosya içeriği. | `pkg/telemetry/qhash.go` |
| 11 | v1 için sadece CLI. Spekülasyonla daemon inşa etme — önce ölç. | (ölçüldü; gerek yok) |
| 12 | `pkg/report`, `pkg/codeindex`'i import etmemeli. | import DAG |
| 13 | Scrub kapalı devre hata verir; yeniden doğrulanamayan not tüm run'ı iptal eder. | `pkg/scrub/scrub.go` |

---

## 13. Latency bütçeleri ve ölçülen değerler

Apple M4 üzerinde ölçülmüş. **Tahmin değil, ölçüm.**

| İşlem | Bütçe | Ölçülen | Not |
|---|---|---|---|
| `forge drift --since-commit` | < 100 ms | **60–70 ms** | **Bağlayıcı kısıt** — git hook yolunda |
| `forge index` | < 200 ms | **20 ms** | |
| `forge check` | < 10 s | **390 ms** sıcak / 930 ms soğuk | |
| `qualitygate.Run` (6 kapı, `code` hariç) | ~ | **~0.13 ms** | |
| `forge verify-code` bash | ~ | ~10 ms sıcak / ~470 ms soğuk | Toolchain başlangıcı baskın |
| `forge verify-code` java | ~ | ~170 ms sıcak / ~370 ms soğuk | |
| `forge verify-code` ts | ~ | **ölçülmedi** | `tsc` bu ortamda kurulu değil |
| `forge session-context` | < 200 ms | bütçenin **çok altında** sıcak | 20 iterasyon, sentetik stdin |
| `forge intent` | < 50 ms | bütçenin **çok altında** sıcak | Sıcak SQLite cache'in yeniden kullanımı |

`forge intent`'in 50 ms bütçesini karşılayabilmesinin **tek** nedeni, `forge recall`'un
zaten sıcak olan SQLite cache'ini yeniden kullanmasıdır. Soğuk bir başlangıçta bu bütçe
mümkün olmazdı — mimari karar (türetilmiş cache) doğrudan bir latency bütçesini mümkün
kılıyor.

**Uyarı:** `hooks/hooks.json` bağlamaları bildirir ama bu ölçümler doğrudan çağrıyla
alındı, canlı bir oturumun ölçümü değil.

---

## 14. Veri katmanının fiziksel yerleşimi

```
<vault>/                          (bir git reposu, örn. ~/Documents/Vault)
├── notes/<type>/<slug>.md        7 tip: concept howto api pattern pitfall incident decision
├── moc/                          Map of Content; moc/weekly/YYYY-WW.md rollup
├── _inbox/                       karantina, confidence: low
├── _archive/
├── profiles/me.md                geliştirici profili (forge init yazar)
├── reports/                      forge check'in dokuz raporu
├── raw/  sources/                not sözleşmesinin dışında, canlı
├── _index.md                     forge index'in çıktısı
└── .forge/
    ├── forge-bin                 vault hook'unun çağıracağı mutlak yol
    ├── capture.log               D3 hook'unun tek teşhis kanalı
    ├── <cache>.db                SQLite — türetilmiş + budget tablosu
    ├── code-index-<repo>.json    repo başına kod indeksi cache'i
    ├── weekly-stats.json         hafta-üstü-hafta delta kalıcılığı
    └── cache/<url-hash>.md       WebFetch kaynak cache'i, TTL'li
```

`pkg/drift/gitindex.go` cache'i kasıtlı olarak repo başına
`.forge/code-index-<repo>.json` olarak yazar, tekil bir `.forge/code-index.json` değil —
`--repo` tekrarlanabilir olduğu için tek bir paylaşılan isim repolar arasında çakışırdı.
`persist`'in doc comment'i bu gerekçeyi taşır.

---

## 15. Test stratejisi

| Katman | Nasıl test edilir |
|---|---|
| Saf fonksiyonlar (`recall`, `similarity`, `graph`, `sentinel`) | Doğrudan birim testi, hiçbir I/O yok |
| Vault işlemleri | `testdata/vault/` fixture'ı geçici dizine kopyalanır, `git init` edilir |
| Drift | Geçici bir git reposu kurulur, commit'lenir, revert edilir (`rollback_test.go`) |
| Engine | `httptest` ile sahte API (`engine_run_httptest_test.go`) |
| CLI | `cmd/forge/e2e_test.go` |
| Cross-lane | `make test`: `CGO_ENABLED=1 go test` **sonra** `CGO_ENABLED=0 go build` |

### `testdata/vault/` fixture'ı

13 not, gerçek vault'un **migration öncesi** topolojisini yeniden üretir, artı **on iki
bilinçli defekt (F1–F12)**: karışık frontmatter şekilleri, sarkan bir wikilink, sarkan
bir `source:` yolu, bir orphan, bir yakın-duplicate çift, hiç frontmatter'ı olmayan
notlar, gövde metninde taşınan status.

İki sert kural:

> **Defektler test yüzeyidir. Onları DÜZELTME.**

> `.git`'i **bilerek yok.** İç içe bir repo, bu repo `git init`-lendiğinde başıboş bir
> gitlink'e dönüşür. Harness fixture'ı geçici bir dizine kopyalar ve **kopyayı**
> `git init` eder. **Asla yerinde `git init` etme.**

Katalog: `testdata/README.md`.

Bu, `examples/vault/` ile karıştırılmamalı — o ayrı bir Phase 6 deliverable'ı: gerçek
vault'tan `forge scrub` ile üretilmiş 93 dosya, sadece `notes/` + `moc/` kapsamında.
