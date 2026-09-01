# Knowledge Forge — Arkasındaki Fikir

> Bu doküman **neden** sorusunu cevaplar. Mimarinin *nasıl*'ı için
> [`02-MIMARI.md`](02-MIMARI.md), kullanımı için [`03-KULLANIM-KILAVUZU.md`](03-KULLANIM-KILAVUZU.md),
> dosya dosya döküm için [`04-DOSYA-DOSYA.md`](04-DOSYA-DOSYA.md).

---

## 1. Çözülmeye çalışılan problem

Bir geliştirici gün içinde onlarca kez şunu yapar:

1. Bir şey takılır — "Hibernate `saveAndFlush` neden `null` timestamp yazıyor?"
2. Claude'a / Google'a sorar.
3. Uzun, iyi bir cevap alır.
4. Sorunu çözer.
5. **Cevabı kaybeder.**

Üç ay sonra aynı soruyu tekrar sorar. Aynı araştırmayı tekrar yaptırır. Aynı token'ları
tekrar harcar. Bu, bilgi işçiliğinin en pahalı israfıdır: **çözülmüş bir problemi
yeniden çözmek.**

Klasik çözüm "not al" demektir. Ama not almanın pratikte üç ölüm nedeni vardır:

| Ölüm nedeni | Ne olur |
|---|---|
| **Sürtünme** | Notu yazmak, cevabı okumaktan uzun sürer. Kimse yapmaz. |
| **Bulunamama** | Not alınır ama altı ay sonra aranmaz — çünkü aramak, yeniden sormaktan yavaştır. |
| **Çürüme (drift)** | Not doğruydu. Kod değişti. Not artık yalan söylüyor ve bunu kimse bilmiyor. |

Knowledge Forge bu üç ölüm nedeninin **üçünü birden** hedefler. Sadece birini çözen bir
araç işe yaramaz: sürtünmesiz ama çürüyen bir not deposu, aktif olarak zarar verir.

---

## 2. v1: `til-writer` skill'i ve sınırları

Bu proje sıfırdan doğmadı. Öncesinde çalışan bir v1 vardı ve hâlâ duruyor:
`~/.claude/skills/til-writer/` — içinde **sadece `SKILL.md`**. Script yok, agent yok,
hook yok, plugin manifest yok.

Kullanıcının global `~/.claude/CLAUDE.md`'si "explain X" tipi promptları bu skill'e
yönlendiriyor, skill de `<vault>/TIL/<konu>/` altına markdown yazıyor. Bu, **sürtünme**
problemini çözdü: not yazmak artık bedava, arka planda oluyor.

Ama diğer ikisini çözmedi:

- **Bulunamama:** 91 not birikti, hiçbiri geri okunmuyordu. Yeni bir soru geldiğinde
  sistem "bu zaten notlarında var" demiyordu. Araştırma her seferinde sıfırdan başlıyordu.
- **Çürüme:** Notlar koda atıf yapıyordu (`repo:path#Symbol`) ama o kod değiştiğinde
  hiçbir şey olmuyordu. Notun `confidence: high` etiketi, altındaki metot silinmiş olsa
  bile `high` kalıyordu.

Ayrıca ölçülebilir kalite problemleri vardı: Phase 1'in migration'ı 91 notu taşıdığında
**60/91 şema-geçerliydi**; kalan 31 notta insan yargısı gerektiren 47 sorun vardı. Yani
"LLM'e yazdır, klasöre at" yaklaşımı, denetimsiz bırakıldığında yarı bozuk bir depo
üretiyor.

**Knowledge Forge, v1'in yerini alan sistemdir** — aynı vault'a yazar, aynı tetikleyiciyi
kullanır, ama arkasına bir statik analiz motoru koyar.

---

## 3. Beş tez

Projenin tamamı beş iddiadan türer. Her mimari karar bunlardan birine dayanır.

### Tez 1 — Araştırmadan önce hatırlama (retrieval-before-research)

Bir soru geldiğinde yapılacak **ilk** şey web araması ya da LLM çağrısı değil, **vault'u
taramaktır.** Cevap zaten yazılmışsa, en pahalı iş (araştırma) hiç başlamamalıdır.

Bu, `forge recall`'un varlık sebebi. Ve kritik detay: recall **hiç model çağrısı
yapmaz.** Eğer "notlarımda bu var mı?" sorusunu cevaplamak için bir LLM çağırmak
gerekiyorsa, tasarruf sıfırlanır — pahalı işten kaçınmak için pahalı iş yapmış olursunuz.

Recall üç eşikli bir karar ağacı döndürür (DESIGN §5.3):

- **≥ 0.85** → `reuse` — not var, aynen kullan, araştırma yapma.
- **0.55 – 0.85** → `update` — ilgili not var, üstüne yaz, sıfırdan başlama.
- **< 0.55** → `create` — gerçekten yeni bir şey, tam pipeline çalışsın.

Verdict'in kendisi `forge recall`'un JSON zarfının içinde döner. Bu bilinçli: eşik ağacını
downstream'de kimse yeniden ifade etmesin, yoksa iki yerde iki farklı eşik olur.

### Tez 2 — Çekirdek sıfır model çağrısı yapar

Bu projenin **en sert invariant'ı**:

> T0 statik çekirdek **sıfır model çağrısı** yapar. Bir tasarım model çağrısı
> gerektiriyor gibi görünüyorsa, dur ve sor.

`forge` binary'sindeki 20 komuttan 19'u tamamen deterministiktir. Tek istisna
`forge engine` — ve o da adı üstünde bir *engine layer*, iş mantığı değil.

Neden bu kadar katı?

1. **Determinizm.** Aynı vault + aynı soru = aynı skor, her zaman. Test edilebilir,
   ölçülebilir, kalibre edilebilir. LLM sokarsanız hiçbiri mümkün değil.
2. **Hız.** `forge drift` 60–70 ms'de biter — çünkü git object store okuyup AST
   karşılaştırıyor, kimseye sormuyor. Bir LLM çağrısı bunu 2 saniye yapardı ve
   **git hook'unun içinde** çalışıyor. 2 saniyelik bir post-commit hook'u kullanıcı
   iki gün içinde siler.
3. **Maliyet.** Statik analiz bedava. Vault her hafta baştan taranabilir çünkü taramanın
   marjinal maliyeti sıfır.
4. **Savunulabilirlik.** LLM sarmalayıcı yazmak kolaydır ve herkes yazar. Bir vault'un
   notları ile bir kod tabanının AST'si arasında git-anchored bir bağ kurmak zordur.
   **Projenin savunulabilir çekirdeği (defensible core) burasıdır**, LLM katmanı değil.

Bunun somut zorlaması: üç pipeline stage'i — `recall`, `write`, `index` — config'de
`none` dışında bir engine kabul **etmez**. Config aksini söylüyorsa binary
**net bir hatayla başlamayı reddeder**, sessizce override etmez. Bu kontrol iki katmanda
var (`pkg/config/validate.go` yükleme anında, `pkg/engine/select.go` seçim anında) ve
ikincisi `engine` alanının yanı sıra `fallback` ve `then` alanlarına da bakar — çünkü
`pipeline.write.fallback: api` yazan bir tamper, `.engine`'e bakan bir kontrolün altından
geçer.

### Tez 3 — Notlar koda bağlıdır, dolayısıyla çürüme tespit edilebilir

Bir not şöyle bir alan taşır:

```yaml
code_refs:
  - myrepo:src/main/java/com/x/PaymentService.java#processRefund
```

Bu sadece bir link değil, **doğrulanabilir bir iddiadır.** O sembol silinirse ya da
imzası değişirse, not artık yanlış olabilir.

`pkg/drift` bunu her commit'te kontrol eder ve beş verdict'ten birini verir:

| Verdict | Anlamı |
|---|---|
| `OK` | Atıf hâlâ çözülüyor, ilgili kod değişmemiş. |
| `Repaired` | Dosya taşınmış ama sembol bulundu — atıf otomatik güncellendi. |
| `Suspect` | İlgili kod değişti; not yanlış *olabilir*. |
| `Broken` | Atıf artık çözülmüyor. **Tek demote eden verdict budur.** |
| `Skipped` | Karar verilemedi (ör. derin tarama gerekiyor). |

Tasarımın en ince yeri **simetri**dir:

> Verdict'ler (not atıfları, ağaç durumu) ikilisinin **saf bir fonksiyonudur**, böylece
> bir revert, demote edilmiş notları simetrik biçimde geri yükler.

Yani `git revert` yaparsanız, notunuz `low`'dan `high`'a kendiliğinden geri döner. Bunun
mümkün olması için `.forge/`'un sakladığı **tek** şey, notun demote'tan *önceki*
confidence değeridir — bir geri yükleme hedefi, asla bir verdict girdisi. Verdict her
seferinde ağaçtan yeniden hesaplanır.

İkinci ince yer: drift **git-anchored**'dır. Dosya kaydında (on save) çalışmaz, çalışma
ağacına (working tree) bakmaz. Sadece `post-commit` / `post-merge` / `post-checkout` ve
`--since-commit <sha>`. Nedeni basit: yarı yazılmış bir dosya bir iddia değildir. Commit
bir iddiadır.

### Tez 4 — Kalite yazma anında zorlanır, sonradan denetlenmez

91 nottan 31'inin bozuk çıkması, "önce yaz, sonra temizleriz" yaklaşımının çalışmadığını
gösterdi. Temizlik hiç gelmez.

Bu yüzden yedi kapı (DESIGN §12) her taslağın önünde durur:

1. **schema** — frontmatter `references/schema.yaml`'e uyuyor mu?
2. **citation** — iddialar kaynaklı mı?
3. **code** — kod örnekleri gerçekten derleniyor mu?
4. **freshness** — not tipine göre tazelik penceresi aşılmış mı?
5. **antislop** — LLM dolgu dili var mı?
6. **link** — wikilink'ler bir yere gidiyor mu?
7. **duplicate** — bu not zaten var mı?

Her kapının bir **remedy**'si vardır ve remedy kapının kendisinden daha önemlidir:

| Kapı | Remedy | Yazmayı bloklar mı? |
|---|---|---|
| schema | `RetryOnce` | evet |
| citation | `MarkUnverified` | evet |
| code / freshness | `DropConfidence` | evet |
| antislop | `RewritePass` | evet |
| link | `DelegateToLibrarian` | **hayır** |
| duplicate | `SwitchToUpdate` | **hayır** |

Son ikisi bloklamaz çünkü onların cevabı "bu notu yazma" değil, "bu notu **farklı**
yaz" — birleştir, ya da bir kütüphaneciye devret.

Ve blocking bir hata olduğunda not **silinmez**, `_inbox/`'a `confidence: low` ile
karantinaya alınır. İlgili invariant:

> Vault'u asla bir zamanlamayla otomatik mutasyona uğratma. Kalite kapısı hataları
> `_inbox/`'a `confidence: low` ile gider, asla sessiz bir publish'e.

"Sessiz publish" burada gerçek tehlikedir: kötü bir not, hiç not olmamasından kötüdür,
çünkü aranır ve bulunur.

### Tez 5 — Markdown tek doğruluk kaynağıdır

SQLite var (`modernc.org/sqlite`, saf Go). Ama:

> Markdown tek doğruluk kaynağıdır. SQLite türetilmiş bir cache'tir; `forge reindex`
> onu tamamen markdown'dan yeniden inşa edebilmelidir.

`forge reindex` bir bakım komutu değil, bir **ispat**tır. Cache'i silip markdown'dan
yeniden kurabiliyorsanız, cache'te sadece markdown'da olan bilgi vardır. Kurulamıyorsa,
bir yere kaçak state sızmıştır.

Tek istisna, bilinçli ve isimlendirilmiş: **budget tablosu.** Günlük harcama sayacı
markdown'da yaşamaz (yaşamamalı da) ve `forge reindex`'ten sağ çıkmalıdır — yoksa
reindex, bütçe sıfırlamanın bir yolu olur.

Bunun pratik faydası: vault bir git reposudur. Notlarınızı `grep`'leyebilir,
Obsidian'da açabilir, elle düzenleyebilir, başka bir makineye `rsync`'leyebilirsiniz.
Araç ölürse veri kalır. Vendor lock-in yoktur çünkü vendor formatı yoktur.

---

## 4. Reddedilen alternatifler

Bir tasarımı anlamanın en hızlı yolu, **seçilmeyeni** ve nedenini okumaktır.

### 4.1 Embedding kullanılmadı (ADR-0001)

`pkg/similarity` elle yazılmış **MinHash + LSH banding**'dir. İlgili invariant tek
cümledir: **"No embeddings."**

Gerekçe:

- Embedding, bir model çağrısı ya da yerel bir model demektir. İkisi de Tez 2'yi kırar.
- Embedding, bir vektör deposu demektir — Tez 5'i kırar, çünkü artık markdown'dan
  yeniden kurulamayan bir state doğar (ya da her reindex'te tüm vault'u yeniden
  embed'lemeniz gerekir).
- Ölçüldü ve **yetti.** Gerçek vault üzerinde 3 duplicate çifti ≥ 0.40 eşiğinde
  bulundu; 140 atıf üzerinden drift 9 notu işaretledi. Bunlar leksikal yöntemle
  bulunan gerçek sonuçlar.

Leksikal skorlama ucuzdur, açıklanabilirdir (`forge recall --explain` size hangi terimin
ne kadar katkı yaptığını gösterir) ve deterministiktir. Bir embedding modeli bunların
üçünü de kaybettirirdi.

### 4.2 Skorlama formülü DESIGN §8'e sadık kalmadı — bilerek

İki karar, ölçülmüş vault davranışından türetildi ve `references/recall-spec.md`
okunmadan geri alınmamalıdır:

**(a) Skor, aktif kanallar üzerinden ağırlıklı bir *ortalamadır*, DESIGN §8'in literal
ağırlıklı toplamı değil.** Dört kanal var — title (0.4), tags (0.3), stack (0.2),
body (0.1). Bir kanal inaktifse (ör. notta `stack` alanı yok) hem paydan **hem de
paydadan** düşer:

```go
for _, c := range chs {
    if !c.Active { continue }
    num += c.Weight * c.Value
    den += c.Weight
}
return num / den
```

Gerekçe kaynakta bir cümleyle yazılı: *"absence of evidence, not evidence of absence"* —
kanıtın yokluğu, yokluğun kanıtı değildir. Notun `stack` alanı boşsa, bu notun eşleşmediği
anlamına gelmez; o kanalın bir şey söylemediği anlamına gelir. Literal toplamda o boşluk
bir 0 gibi davranır ve notu haksızca cezalandırır.

**(b) Başlık ölçüsü Dice değil, F₂'dir.** Bu, iki gerçek vakadan çıktı:

- Saf coverage (recall) kullanınca *"Spring Boot 4 Breaking Changes"* notu
  *"how does spring boot work"* sorusuna **mükemmel** puan verdi — çünkü sorunun tüm
  terimleri başlıkta geçiyordu, ama not tamamen başka bir şeyle ilgiliydi.
- Simetrik Dice kullanınca *"Keyset Pagination — Compound OR Predicate"* notu 0.67'ye
  düştü — uzun başlık cezalandırıldı, oysa not doğru cevaptı.

F₂ ile bu ikisi **0.59** ve **0.83** oldu. Yani recall'a precision'ın 2 katı ağırlık
vermek, her iki patolojiyi aynı anda düzeltti:

```go
p := float64(hits) / float64(titleTokens)
r := float64(hits) / float64(queryTerms)
return 5 * p * r / (4*p + r)
```

Benzer şekilde IDF, `log(n/df)` değil **`log(1 + n/df)`** — çünkü düzeltilmemiş biçim,
bir terim *her* notta geçtiğinde tam olarak sıfır verir ve terim denklemden tamamen
düşer. Üst sınır `idfCap = 3.5`.

### 4.3 Python değil Go (ADR-0002)

ADDENDUM §B başlangıçta Python öngörüyordu. `docs/KNOWLEDGE-FORGE-STACK.md` (ADR-001) bunu
tersine çevirdi ve dokümanın kendisi *"that was wrong"* diyor. Nedenler:

- **Tek dosya dağıtım.** `forge` tek bir statik binary. Python bir runtime, bir venv ve
  bir bağımlılık ağacı demektir — bir git hook'unun içinde bunların hepsi kırılganlıktır.
- **Latency bütçesi.** 100 ms'lik drift bütçesi, Python interpreter başlangıcının
  neredeyse tamamını yerdi.
- **Cross-compile.** Makefile dört hedefe derliyor (darwin/linux × amd64/arm64)
  tek komutla, `CGO_ENABLED=0` sayesinde. Windows şimdilik hedef değil.

Hayatta kalan tek Python: bir kerelik `migrate_vault.py` ve offline dataset/fine-tuning
araçları. İkisi de binary'ye girmez.

### 4.4 Daemon yok

> v1 için sadece CLI. Spekülasyonla daemon inşa etme — önce ölç.

Ve ölçüldü: `forge check` sıcak **390 ms** (soğuk 930 ms), bütçe 10 saniyeydi.
`forge index` **20 ms**, bütçe 200 ms. Ölçülen rakamlar bütçenin çok altında olunca,
"sürekli çalışan bir servis" için hiçbir gerekçe kalmadı. Daemon, çözülmemiş bir
problemin çözümü olurdu.

---

## 5. LLM katmanı: neden var, neden yukarıda

Sıfır model çağrısı, "LLM yok" demek değil. "LLM **çekirdekte** yok" demek.

Dört katman (tier) var:

| Tier | Ne | Maliyet |
|---|---|---|
| `none` | Model yok. Saf statik. | 0 |
| `host` | Zaten oturumda olan Claude Code. | 0 (kullanıcının aboneliği) |
| `api` | Anthropic API (varsayılan `claude-sonnet-5`). | Ölçülür |
| `advisor` | Daha güçlü model (`claude-opus-5`), **sadece kritik**. | Ölçülür |

Üç fikir bunu bir arada tutuyor:

**(1) Katman başına değil, *stage* başına yönlendirme.** Pipeline'ın dokuz aşamasının
her biri kendi engine'ini, kendi `fallback`'ini ve kendi `then`'ini seçer. Örnek shipped
config:

```yaml
recall:     { engine: none }          # LOCKED
research:   { engine: api, fallback: host }
verify:     { engine: advisor, fallback: local, then: host }
write:      { engine: none }          # LOCKED
```

Yani araştırma pahalı bir modele gidebilir, ama *yazma* asla gitmez. Ucuz olan iş ucuz
kalır.

**(2) Bütçe gerçek ve zorlayıcıdır.** `advisor_usd_per_day: 2.00`,
`api_usd_per_day: 1.00`. Sayaçlar SQLite'ta yaşar ve `forge reindex`'ten sağ çıkar.
Bütçe bittiğinde davranış `on_exhausted: queue` — not kuyruğa alınır, sessizce düşük
kaliteli bir sonuç üretilmez.

Burada ince bir ayrım var: `Exhausted()` fonksiyonu, "bugün bütçe yok" ile "buraya zaten
ölçülen bir tier hiç konfigüre edilmemiş" arasını ayırt eder. `Resolve()` tek bir kazanan
isim döndürdüğü için bu bilgiyi kaybeder — `queue` kararı vermeden önce bu ayrım şart,
yoksa `engine: none` olan bir stage sonsuza kadar kuyruğa yazardı.

**(3) Advisor sadece eleştirir.**

> Advisor tier'ı (T3) yalnızca kritik yapar: tartışmalı iddiaları ve bir patch döndürür,
> asla bir yeniden yazım.

Bu ekonomik olduğu kadar epistemolojik bir karardır. Pahalı modelin işi "daha güzel yaz"
değil, "**burada yanlış olan ne?**"dir. Bir yeniden yazım, ucuz modelin doğru yaptığı
şeyleri de siler ve neyin neden değiştiğini görünmez kılar. Bir patch denetlenebilir.

---

## 6. Bileşik döngü (compounding loop)

Parçalar tek tek faydalıdır, ama asıl fikir **birbirlerini beslemeleridir**:

```
  soru sorulur
       │
       ▼
  ┌─ recall ──────► not zaten var mı?  ──evet──► reuse (araştırma yok)
  │                         │hayır
  │                         ▼
  │                   research → synthesize → gate → yaz
  │                         │
  │                         ▼
  │                   not, code_refs ile koda bağlanır
  │                         │
  ▼                         ▼
 index ◄──────────── vault büyür
  │                         │
  │                         ▼
  │                   kod commit'lenir → drift hook → verdict
  │                         │
  │                         ▼
  │                   çürüyen not demote edilir (revert ederse geri döner)
  │                         │
  ▼                         ▼
 logback ────────► kod reposu kendi CLAUDE.md'sinde notları görür
                            │
                            ▼
                   bir sonraki soru, daha zengin bir vault'a çarpar
```

Her tur döngüyü güçlendirir:

- **capture** (`forge capture`, D3 hook) vault commit'lerinden insan-düzeltmesi
  çiftleri toplar — ileride fine-tuning için dataset.
- **recall** her yeni notla daha isabetli olur.
- **drift** her yeni `code_refs` ile daha çok yüzey kaplar.
- **logback** bilgiyi vault'tan kod reposuna geri taşır: `docs/knowledge-map.md`,
  modül başına `CLAUDE.md` fragment'leri ve opsiyonel satır içi
  `// forge:logback:<symbol>` işaretleri. Böylece vault'u hiç açmayan bir geliştirici
  bile, üzerinde çalıştığı dosyanın hangi notlarla ilişkili olduğunu görür.
- **stats** (`forge stats`) döngünün kendisini ölçer: hit rate, en çok sorulan konular,
  boşluklar, yaklaşık kazanılan zaman.

`logback`'in tersine akışı özellikle önemlidir. Bilgi tabanlarının klasik ölüm biçimi,
"ayrı bir yerde durup kimsenin uğramamasıdır". Logback, bilgiyi geliştiricinin *zaten*
baktığı yere — kod reposunun içine — enjekte eder. Ve bunu yaparken `pkg/sentinel`'in
id'li begin/end blok primitifini kullanır: **kendi işaret çiftinin dışındaki hiçbir
şeye, tek bayt bile dokunmaz.** İdempotenttir, konumdan bağımsızdır, ve
`--remove-markers` ile bayt-bayt geri alınabilir.

---

## 7. Gizlilik ve güvenlik duruşu

Bilgi tabanı, tanımı gereği hassas şeyler biriktirir. Üç mekanizma bunu ele alır:

**Telemetri minimumdur.**

> Telemetri konuyu ve bir hash'i loglar. Asla ham soru metnini, kodu ya da dosya
> içeriğini.

`pkg/telemetry` sha256 ile konu hash'ler ve tamamı `cfg.Telemetry.Enabled`'ın arkasında
kapalıdır.

**Kod doğrulama izole edilir.**

> Kod doğrulama tek kullanımlık bir dizinde derler, asla kullanıcının projesinde.

`forge verify-code` sistem toolchain'ini çağırır ama bunu geçici bir alanda yapar —
bir not içindeki kod örneğini derlemek, kullanıcının build'ini kirletmemelidir. Ve bu bir
bağımlılık çözücü değildir; sadece "bu snippet sentaktik/tipsel olarak ayakta mı?"
sorusunu cevaplar.

**Dışa aktarım kapalı devre hata verir (fails closed).**

`pkg/scrub` bir vault kopyasından e-posta, mutlak ev dizini yolu, anahtar önekli token ve
uzun token biçimli içeriği redakte eder. Kritik davranış:

> Scrub'dan sonra `references/schema.yaml`'e karşı yeniden doğrulanamayan bir not,
> **tüm çalışmayı iptal eder** ve `--dst`'ye hiçbir şey yazılmaz.

Bu, `file{rel, data}` yapılarının her not başarılı olana kadar bellekte tutulmasıyla
sağlanır. Yarı redakte edilmiş bir çıktı, hiç çıktıdan tehlikelidir — kullanıcı onu
temiz sanır.

Scrub'ın heuristiğinin gerçek vault üzerinde nasıl kalibre edildiği öğreticidir. İlk
sürüm "32+ karakterlik herhangi bir `[A-Za-z0-9]` dizisi" arıyordu ve **637 redaksiyon**
üretti. İki gerçek false-positive sınıfı bulundu:

1. **kebab-case slug'lar ve tarihli dosya adları** (`2026-04-13-local-ai-continue-rag-spring`)
   — `sources:` ve wikilink atıflarını bozuyordu. Düzeltme: `-` ve `_` karakter
   sınıfından çıkarıldı. → **86 redaksiyon**
2. **camelCase Java tanımlayıcıları** (`getPaymentOutboxMessageBySagaIdAndSagaStatus`)
   — gömülü kod örneklerinde. Düzeltme: eşleşmede en az bir rakam zorunlu kılındı.
   RE2'de lookahead olmadığı için filtre eşleşme sonrası çalışır. → **43 redaksiyon**

İkinci düzeltmenin güvenlik gerekçesi açıktır: `[A-Za-z0-9]`'dan rastgele çekilen 32+
karakterlik bir dizi neredeyse kesinlikle bir rakam içerir, dolayısıyla gerçek
hex/base64/JWT biçimli sırlar hâlâ yakalanır. Her iki düzeltmenin de regression testi
var. Kovalanmayan bir artık false-positive kaldı: içinde rakam barındıran bir
tanımlayıcı (ör. `TestE2ESessionContextRespectsTheBudget`, "E2E" yüzünden) hâlâ tetikler.

---

## 8. Neden bir Claude Code *plugin*'i

Araç, kullanıcının zaten bulunduğu yerde yaşamazsa kullanılmaz. Bu yüzden Knowledge Forge
bağımsız bir CLI olarak değil (öyle de çalışsa) bir plugin olarak paketlenir:

- **Dört lifecycle hook'u** oturuma bağlanır: `SessionStart` vault index'ini + geliştirici
  profilini context'e basar; `UserPromptSubmit` her promptta ucuz bir recall kontrolü
  yapar ve 0.7 üstündeki hit'i `additionalContext` olarak ekler; `SessionEnd`
  transcript'ten sonuç cümlelerini tarayıp `_inbox/`'a stub yazar; `PostToolUse`
  (`WebFetch`) çekilen kaynağı TTL'li olarak cache'ler.
- **Dört skill** komutları sarar (`forge`, `forge-init`, `forge-check`, `forge-stats`).
- **Dört product agent** spesifikasyonu araştırma/doğrulama/kütüphanecilik işini böler.

Hook'ların hepsi **fail-silent**'tır ve **her zaman 0 döner.** Bu tartışılmaz: bir
bilgi tabanı aracı, kullanıcının commit'ini ya da oturumunu asla kıramaz. Vault'un D3
hook'u da aynı sözleşmeyle çalışır — "tasarım gereği hook bir commit'i asla
başarısız edemez ve asla bir şey basmaz". Bunun bedeli, bozuk bir binary'nin **sessiz**
olmasıdır; teşhis için `<vault>/.forge/capture.log` vardır.

Hook'larla ilgili, kaynakta belgelenmiş bir tuzak: `--continue`/`--resume` ile devam eden
bir oturumda `SessionStart` yeniden çalışır (beklenen; çıktısı idempotent ve ucuz), ama
**diğer her hook'un çıktısı kaydedilmiş transcript'ten tekrar oynatılır**, yeniden
çalıştırılmaz. Yani resume'da eski bir recall hit'i görmek bir bug değil, beklenen
davranıştır — ve bu yüzden hiçbir hook'a zamana duyarlı bir iş konmamalıdır.

---

## 9. Özet: tek cümlelik fikir

> **Bir geliştiricinin "şunu açıkla" anlarını, koda git üzerinden bağlı ve bu yüzden
> çürüdüğünde bunu kendisi söyleyebilen kalıcı markdown notlarına çeviren, çekirdeği
> sıfır model çağrısı yapan bir statik analiz motoru.**

LLM katmanı isteğe bağlıdır ve yukarıdadır. Değeri üreten şey, altındaki deterministik
motordur.
