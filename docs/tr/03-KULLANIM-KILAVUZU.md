# Knowledge Forge — Kullanım Kılavuzu

> *Neden* için [`01-FIKIR.md`](01-FIKIR.md), *nasıl çalışır* için
> [`02-MIMARI.md`](02-MIMARI.md), dosya dökümü için [`04-DOSYA-DOSYA.md`](04-DOSYA-DOSYA.md).

---

## 0. Hızlı başlangıç

```bash
# 1. Derle (saf Go lane — dağıtılan lane budur)
cd /Users/mimir45/TIL
CGO_ENABLED=0 go build ./...
make build                       # → ./dist/forge

# 2. Kod indeksleme (tree-sitter) istiyorsan cgo lane'i
make full                        # CGO_ENABLED=1, -tags codeindex

# 3. Kurulum sihirbazı
forge init --vault ~/Documents/Base --language java \
           --frameworks spring-boot,hibernate --seniority senior

# 4. Config'in gerçekten ne çözdüğünü gör
forge config --layers
forge config --json

# 5. İlk gerçek kullanım
forge recall --question "how does keyset pagination work" --explain
```

> ⚠️ **Bayat binary tuzağı.** Repo kökündeki checked-in `./forge` **eskidir** —
> Phase 5b ve 6'dan önce derlenmiş, `logback` ve `scrub` komutlarını **tanımaz**
> (`forge: unknown command`). Kök binary'yi kullanmadan önce `make build` çalıştır ya da
> `./dist/forge`'u kullan.

---

## 1. Kurulum

### 1.1 Kaynaktan derleme

| Komut | Ne yapar |
|---|---|
| `make build` | `CGO_ENABLED=0`, kod indeksleme yok, cross-compile edebilir. |
| `make full` | `CGO_ENABLED=1 -tags codeindex`, tree-sitter derlenir, host toolchain gerekir. |
| `make test` | Önce `CGO_ENABLED=1 go test ./...`, sonra `CGO_ENABLED=0 go build ./...`. **İkisi de gerekli.** |
| `make bench` | Benchmark'lar. |
| `make vet` / `make fmt` / `make lint` | `go vet`; `gofmt`; `gofmt`-as-a-gate + `go vet`. |
| `make dist` | Altı hedef: darwin/linux/windows × amd64/arm64. |
| `make checksums` | `bin/forge` shim'inin çalıştırmadan önce doğruladığı sha256'lar. |
| `make install-hook` | Binary'yi `$HOME/.forge/bin/forge`'a kopyalar + `.sha256` pin'i yazar. |
| `make clean` / `make help` | |

### 1.2 Claude Code plugin'i olarak

```bash
claude plugin marketplace add mimir45/Knowledge-Forge
```

> **Doğrulanmamış:** Bu yol ve `bin/forge` shim'inin indir-ve-checksum akışı, tag'lenmiş
> bir release gerektiriyor ve tertemiz bir makinede henüz denenmedi. Bugün için kaynaktan
> derleme yolu güvenilir olandır.

Plugin kurulduğunda `.claude-plugin/plugin.json`, `hooks/hooks.json`'ı ve `agents/`'ı
otomatik keşfeder. `settings.json`'a elle kopyalama gerekmez.

---

## 2. `forge init` — kurulum sihirbazı

**Tam olarak iki dosya yazar, başka hiçbir şey:**

| Dosya | İçerik |
|---|---|
| `~/.forge/forge.config.md` | Sadece paketlenmiş varsayılanlardan **farklı** olan anahtarlar — böylece binary yükseltmesi, karar vermediğin her şey için yeni varsayılanları getirmeye devam eder. |
| `<vault>/profiles/me.md` | Geliştirici profili, `profiles/me.template.md`'den render edilir. |

`config/forge.config.md`'ye **asla** yazmaz — o paketlenmiş bir şablondur.

```
forge init --vault DIR [--language L] [--frameworks a,b] [--infra a,b]
           [--seniority junior|mid|senior] [--depth 1-5] [--note-language en]
           [--explain-style mechanism-first] [--trigger ask|auto|manual]
           [--engine-preset claude-only] [--stack-preset java-backend]
           [--force] [--dry-run]
```

| Flag | Varsayılan | Not |
|---|---|---|
| `--vault` | — | **Zorunlu.** |
| `--language` | — | Birincil dil, ör. `java`. |
| `--frameworks` | — | Virgüllü, ör. `spring-boot,hibernate`. |
| `--infra` | — | Virgüllü, ör. `docker,postgres,kafka`. |
| `--seniority` | `mid` | `junior \| mid \| senior`. |
| `--depth` | `0` | 1–5; 0 ise `--seniority`'den türetilir. |
| `--note-language` | `en` | Not gövdelerinin dili. |
| `--explain-style` | `mechanism-first` | `mechanism-first \| example-first \| analogy-first`. |
| `--trigger` | `ask` | `ask \| auto \| manual`. |
| `--engine-preset` | `claude-only` | `offline \| claude-only \| byo-api \| max`. |
| `--stack-preset` | — | `java-backend \| frontend \| devops \| minimal`. |
| `--force` | false | Var olan dosyaların üzerine yaz. |
| `--dry-run` | false | Ne yazılacağını bas, hiçbir şey yazma. |

**Komut soru sormaz.** Soruları `skills/forge-init/` sorar ve bunu çağırır. `--force`
olmadan var olan bir dosyanın üzerine yazmayı reddeder.

Beş soru: dil, framework'ler, infra, seniority, trigger. `depth` ve assume/never
listeleri türetilir — sihirbaz beşte kalsın diye.

Exit: `3` / `4` — vault ön koşulu sağlanmadığında (bkz. `checkVault`).

---

## 3. Komut referansı

`forge` 20 komut taşır. Ortak desen: neredeyse hepsi `--vault` alır ve boş bırakılırsa
config'in `vault_path`'ini, o da yoksa `.`'yı kullanır.

### 3.1 Config ve keşif

#### `forge config`

```bash
forge config --layers      # katkı veren dosyaları listele
forge config --json        # birleştirilmiş config'i JSON olarak bas
```

Bir "neden bu değer?" sorusunda **ilk** çalıştırılacak komut. Dört katmanı ve hangisinin
kazandığını gösterir.

#### `forge slug`

```bash
forge slug "Keyset Pagination — Compound OR Predicate"
forge slug --vault ~/Documents/Base --json "Some Title"
```

| Flag | Not |
|---|---|
| `--vault` | Verilirse mevcut slug'lardan kaçınılır (çakışma yok). |
| `--json` | `{"title":…,"slug":…}` biçiminde çıktı. |

Kanonik kebab-case slug üretir. Bir not yazmadan önce dosya adını belirlemek için.

### 3.2 Vault bakımı

#### `forge validate`

```bash
forge validate notes/concept/hibernate.md
forge validate --all --vault ~/Documents/Base
forge validate --all --fix --quiet
```

| Flag | Not |
|---|---|
| `--all` | `--vault` altındaki her notu doğrula. |
| `--vault` | Vault kökü. |
| `--fix` | Mekanik olarak düzeltilebilir sorunları yerinde onar. |
| `--quiet` | Sadece özet satırını bas. |

**Exit 0** = her not uyumlu. **Exit 1** = en az biri değil.

`--fix` yalnızca mekanik olanı düzeltir. İnsan yargısı gerektirenlere dokunmaz — Phase
1'in migration'ında 91 nottan 31'i tam bu kategorideydi (47 sorun).

#### `forge index` / `forge reindex`

```bash
forge index --vault ~/Documents/Base
forge index --out _index.md --max-bytes 4096
forge reindex --vault ~/Documents/Base
```

| Flag | Varsayılan | Not |
|---|---|---|
| `--vault` | config → `.` | |
| `--out` | config `paths.index` | Vault köküne göreli. |
| `--max-bytes` | `4096` | Üretilen index için bayt bütçesi. |

`index` `<vault>/_index.md`'yi markdown'dan yeniden kurar. **20 ms** sürer.

`reindex` türetilmiş SQLite cache'ini **atar** ve markdown'dan tamamen yeniden kurar.
Bu bir bakım komutu değil, Tez 5'in **ispatıdır**: cache'te yalnızca markdown'da olan
bilgi var. Tek istisna — **budget tablosu reindex'ten sağ çıkar.**

### 3.3 Recall — araştırmadan önce hatırlama

```bash
forge recall --question "how does keyset pagination work"
forge recall --question "spring boot startup" --stack java,spring-boot --explain
```

| Flag | Not |
|---|---|
| `--vault` | Vault kökü. |
| `--question` | Vault'a karşı eşleştirilecek soru. |
| `--stack` | Virgüllü stack ipuçları, ör. `java,spring-boot`. |
| `--explain` | Aday başına skor dökümünü **stderr**'e bas. |

Çıktı stdout'ta JSON zarfı; **verdict zarfın içinde** gelir:

| Skor | Verdict | Anlamı |
|---|---|---|
| ≥ 0.85 | `reuse` | Not var, aynen kullan. Araştırma yapma. |
| 0.55 – 0.85 | `update` | İlgili not var, üstüne yaz. |
| < 0.55 | `create` | Yeni. Tam pipeline. |

`--explain` çıktısı hangi kanalın (title/tags/stack/body) ne katkı yaptığını, terim
ağırlıklarını ve bayatlık işaretini gösterir. Bir skorun *neden* öyle olduğunu anlamak
için tek yol budur.

**Sıfır model çağrısı.** Bu komut asla bir LLM'e gitmez.

### 3.4 Drift — çürüme tespiti

```bash
# salt-okunur kontrol
forge drift --repo myapp=/path/to/repo --vault ~/Documents/Base

# hook yolu: sadece değişen dosyalar, demote uygula
forge drift --repo myapp=/path/to/repo --since-commit abc1234 --apply

# derin tarama: eksik sembolleri notun doğrulanmış-era revizyonunda ara
forge drift --repo myapp=/path/to/repo --deep --json
```

| Flag | Not |
|---|---|
| `--vault` | Vault kökü. |
| `--repo` | `name=path` biçiminde, **tekrarlanabilir**, **en az bir tane zorunlu**. |
| `--since-commit` | Sadece bu sha'dan beri değişen dosyaları değerlendir; boş bırakınca her atıf kontrol edilir. |
| `--deep` | Her repo'yu notun doğrulanmış-era revizyonunda yeniden indeksle (eksik sembollere karar vermek için). |
| `--apply` | Confidence'ı taşı ve `drift_checked_at` damgası vur. **Bu olmadan run salt-okunurdur.** |
| `--json` | Bulguları JSON olarak bas. |

**Verdict'ler:** `OK`, `Repaired`, `Suspect`, `Broken`, `Skipped`. Sadece `Broken`
demote eder.

**Ölçülen:** `--since-commit` ile **60–70 ms** (bütçe 100 ms). Bu, projenin bağlayıcı
latency kısıtı.

**Simetri garantisi:** `git revert` yaparsanız aynı ağaç → aynı verdict → not
kendiliğinden geri yükselir. `.forge/` sadece demote öncesi confidence'ı saklar.

### 3.5 Check — haftalık geçiş

```bash
forge check --vault ~/Documents/Base --repo myapp=/path/to/repo
forge check --offline                 # ağa çıkma; deadlinks sadece cache raporlar
forge check --months 6 --days 90
```

| Flag | Varsayılan | Not |
|---|---|---|
| `--vault` | config → `.` | |
| `--repo` | — | `name=path`, tekrarlanabilir. |
| `--months` | `0` | `churn.md` için vault geçmişi penceresi; 0 = tamamı. |
| `--days` | `0` | Kod churn penceresi (gün); 0 = config `check.churn_days` (90). |
| `--offline` | false | Ağ atlanır; `deadlinks.md` yalnızca cache'lenmiş verdict'leri raporlar. |

`<vault>/reports/` altına dokuz raporu render eder:

`coverage` · `staleness` · `duplicates` · `orphans` · `gaps` · `graph-health` ·
`churn` · `deadlinks` · `drift`

Ayrıca: `cost.md`, `codebase.md`, `moc/weekly/YYYY-WW.md` rollup'ı ve
`.forge/weekly-stats.json` hafta-üstü-hafta kalıcılığı.

**Deterministik:** altı ardışık çalıştırma md5-özdeş. `writeReport` yalnızca içerik
değiştiyse yazar — değişmemiş bir raporun mtime'ı oynamaz.

**Ölçülen:** 390 ms sıcak / 930 ms soğuk (bütçe 10 s).

Config'de `check.schedule: "0 9 * * MON"` yazar ama bu bir **öneridir** — invariant
gereği hiçbir şey vault'u zamanlamayla otomatik mutasyona uğratmaz.

### 3.6 Engine — dört tier

```bash
forge engine select --stage research --json
forge engine run --stage research --prompt-file /tmp/p.md --rel notes/concept/x.md
forge engine record --stage plan --tier host --rel notes/concept/x.md
```

> `forge engine --help` kozmetik bir tuhaflık gösterir: usage bloğundan önce
> `forge engine: unknown subcommand "--help"` basar. Zararsız.

#### `engine select` — kuru çalıştırma

| Flag | Not |
|---|---|
| `--vault` | |
| `--stage` | Çözülecek pipeline stage'i, ör. `research`. |
| `--json` | Sonucu JSON olarak bas. |

**HTTP yok, harcama yok.** Zinciri (`engine` → `fallback` → `then`) yürür ve kazananı
**gerekçesiyle** döndürür — "offline neden `none`'a düştü" sorusunu cevaplar.

Kilitli bir stage'e (`recall`/`write`/`index`) `none` dışında bir şey yazılmışsa burada
net bir hata verir:

```
engine: pipeline.write: "api" is not allowed — [recall write index] are locked to
"none" (T0 static core)
```

#### `engine run` — gerçek çağrı

| Flag | Not |
|---|---|
| `--vault` | |
| `--stage` | Çalıştırılacak stage. |
| `--prompt-file` | Gönderilecek prompt'u içeren dosya. |
| `--rel` | Notun vault-göreli yolu. |

Çözülen engine'i çağırır ve maliyetini **bugünün bütçesine yazar**. Bütçe bittiyse
`on_exhausted: queue` devreye girer: not kuyruğa alınır, düşük kaliteli bir sonuç
üretilmez. Kuyruk `forge check`'in `drainAdvisorQueue`'suyla boşaltılır.

Bu, binary'nin **gerçek bir model çağrısı yapabilen tek komut ağacıdır.**

#### `engine record` — engine_trail damgası

| Flag | Not |
|---|---|
| `--vault` | |
| `--stage` | Tier'ın çalıştığı stage. |
| `--tier` | `host \| api \| advisor`. |
| `--rel` | Notun vault-göreli yolu. |

Host-tier bir adım tamamlandıktan sonra notun frontmatter'ına `engine_trail` basar.
**Kilitli bir stage'e kayıt yapmayı reddeder.**

### 3.7 Kalite

#### `forge gate` — yedi kapı

```bash
forge gate --file /tmp/draft.md --rel notes/concept/x.md
forge gate --file /tmp/draft.md --rel notes/concept/x.md \
           --mode update --target-slug hibernate-flush-order
```

| Flag | Varsayılan | Not |
|---|---|---|
| `--file` | — | Render edilmiş taslak notun yolu. |
| `--rel` | — | Notun hedeflenen vault-göreli yolu. |
| `--vault` | config → `.` | |
| `--mode` | `create` | `create` veya `update`. |
| `--target-slug` | — | `update` modunda: genişletilen notun slug'ı. |
| `--previous-draft` | — | Önceki bir karantinadan gelen yol; D4 için eşleştirmeye. |

Sıra: `schema → citation → code → freshness → antislop → link → duplicate`.

| Exit | Anlamı |
|---|---|
| **0** | `Quarantine false` — temiz yayınlandı. |
| **1** | `Quarantine true` — **hata değil**; not doğru işlendi, sadece yayınlanmadı. |
| **2** | Kullanım hatası. |
| **3** | İç hata: kapı çalıştırma ya da karantina yazımının kendisi başarısız. |

Exit 1'in "hata değil" oluşu önemlidir: karantina **beklenen** bir sonuçtur, bir
başarısızlık değil. Bir CI script'i exit 1'i kırmızıya boyamamalı.

#### `forge verify-code`

```bash
forge verify-code --lang java --file /tmp/Snippet.java
cat snippet.sh | forge verify-code --lang bash --stdin
forge verify-code --lang auto --file /tmp/x.ts
```

| Flag | Not |
|---|---|
| `--lang` | `java`, `ts`, `bash` veya `auto` (`auto` için `--file` gerekir). |
| `--file` | Snippet'in yolu; `--stdin` okumak için boş bırak. |
| `--stdin` | Snippet'i stdin'den oku. |

Makinede **zaten kurulu olan** `javac`/`tsc`/`bash`'i kullanır. **Bir bağımlılık
çözücü değildir.** Tek kullanımlık bir dizinde derler, asla kullanıcının projesinde.

**Exit 0** = geçti veya atlandı, **1** = başarısız.

Ölçülen: bash ~10 ms sıcak, java ~170 ms sıcak. Maliyet kapı mantığı değil, toolchain
başlangıcı. `tsc` bu ortamda kurulu değil.

### 3.8 Logback — bilgiyi kod reposuna geri taşı

```bash
forge logback --repo myapp=/path/to/repo --dry-run
forge logback --repo myapp=/path/to/repo
forge logback --repo myapp=/path/to/repo --remove-markers
```

| Flag | Not |
|---|---|
| `--vault` | |
| `--repo` | `name=path`, tekrarlanabilir, **zorunlu**. |
| `--dry-run` | Ne değişeceğini bas, hiçbir şey yazma. |
| `--remove-markers` | Bu komutun yazdığı satır içi `// forge:` işaretlerini sil; **logback config kapılarını yok sayar**. |

Üç çıktı, her biri **bağımsız olarak** config'le kapılanır:

| Çıktı | Config anahtarı | Varsayılan |
|---|---|---|
| `docs/knowledge-map.md` | `static.logback.knowledge_map` | açık |
| Modül başına `CLAUDE.md` fragment'leri | `static.logback.claude_md_fragment` | açık |
| Satır içi `// forge:logback:<symbol>` işaretleri | `static.logback.inline_markers` | **kapalı** |

T0, deterministik, **idempotent**. Byte-özdeş yeniden çalıştırma doğrulandı (`diff`,
çıktı yok) ve `--remove-markers` byte-bayt geri alma doğrulandı.

`pkg/sentinel` sayesinde **kendi işaret çiftinin dışındaki hiçbir şeye dokunmaz.**

> **Uygulama notu (doğruluk düzeltmesi):** satır içi işaret çözümlemesi
> `coderef.Ref.Symbol != ""`'e bakmalı, `Ref.Kind == KindSymbol`'e değil. Kanonik
> `code_refs:` biçimi (`repo:path#Symbol`) `KindPath` + `Symbol` set olarak parse olur;
> sadece `Kind`'a bakmak neredeyse her gerçek atıfı sessizce atlardı.

### 3.9 Scrub — güvenli dışa aktarım

```bash
forge scrub --src ~/Documents/Base --dst /tmp/vault-clean
```

| Flag | Not |
|---|---|
| `--src` | Kaynak vault dizini. |
| `--dst` | Redakte edilmiş kopyanın hedefi. |

`--src`'yi bir vault olarak yürür; her notun frontmatter ve gövdesinden sır/PII biçimli
içeriği (e-posta, mutlak ev yolu, API-anahtarı biçimli token) redakte eder ve aynı göreli
düzenle `--dst` altına yazar.

**Kapalı devre hata verir (fails closed):** herhangi bir not scrub'lanıp
`references/schema.yaml`'e karşı yeniden doğrulanamıyorsa, `--dst`'ye **hiçbir şey**
yazılmaz — temiz scrub'lanmış notlar bile. Hata durumunda `--dst`'ye asla dokunulmaz.

Başarıda stdout'a JSON `Report`: `NotesTotal`, `NotesWritten`, `Redactions`,
`NoFrontmatter` (gövde-only yazılan, hâlâ frontmatter'ı olmayan migration-öncesi notlar).

| Exit | Anlamı |
|---|---|
| **0** | Başarı. |
| **2** | Kullanım hatası. |
| **3** | Scrub başarısız, `--dst` dokunulmamış. |

`examples/vault/`'u üretmek için kullanıldı ve Phase 6b'nin `--anonymize` export yolu
bunu çağıracak.

### 3.10 Dataset yakalama

#### `forge capture`

```bash
forge capture --vault ~/Documents/Base --commit HEAD --dry-run
```

| Flag | Varsayılan | Not |
|---|---|---|
| `--vault` | config → `.` | Vault **git reposu**. |
| `--commit` | `HEAD` | Hasat edilecek commit. |
| `--window-days` | `7` | Üretim ile düzenleme arasındaki maksimum gün. |
| `--out` | `dataset.D3Path` | Dataset dosyası, vault köküne göreli. |
| `--dry-run` | false | Çiftleri yazmadan raporla. |
| `--quiet` | false | Hiç çift yakalanmadıysa hiçbir şey basma. |

Bir vault commit'inden **insan-düzeltmesi** eğitim çiftleri hasat eder. Vault'un D3
post-commit hook'u tarafından çağrılır.

> **`Forge-Write: true` neden önemli:** `forge-librarian` agent'ı authored ettiği her
> commit'e bu damgayı basar. Basmasaydı, `pkg/dataset` agent'ın kendi çıktısını
> *insan düzeltmesi* olarak kaydederdi ve eğitim verisi kirlenirdi.

### 3.11 Stats

```bash
forge stats --vault ~/Documents/Base
```

Beş bölüm basar:

| Bölüm | Ne gösterir |
|---|---|
| Hit rate | Sorulan sorulardan kaçı vault'ta karşılık buldu. |
| Top topics | En çok sorulan konular. |
| Gaps | Sorulmuş ama yazılmamış konular. |
| Time saved | Yaklaşık kazanılan zaman tahmini. |
| Trend | `.forge/weekly-stats.json`'dan hafta-üstü-hafta değişim. |

Bu, döngünün kendisini ölçen komuttur. "Gaps" bölümü doğrudan bir yazma listesidir.

### 3.12 Hook komutları

Bu dördü elle çalıştırılmak için değil, Claude Code lifecycle'ına bağlanmak içindir.
Hepsi **fail-silent** ve **her zaman exit 0**.

| Komut | Event | Ne yapar | Flag'ler |
|---|---|---|---|
| `forge session-context` | `SessionStart` | Vault index'ini + geliştirici profilini context'e basar. | `--vault`, `--max-bytes` (4096; index'e ve profile **ayrı ayrı** uygulanır) |
| `forge intent` | `UserPromptSubmit` | stdin'deki prompt'a ucuz recall kontrolü; **0.7 üstündeki** en iyi hit'i `additionalContext` olarak yayar. | `--vault` |
| `forge session-capture` | `SessionEnd` | stdin'deki transcript'i sonuç cümleleri için tarar, `_inbox/`'a **en fazla 3** düşük-confidence stub yazar; session-id + içerik hash'iyle dedupe. | `--vault` |
| `forge cache-source` | `PostToolUse` (`WebFetch`) | `.forge/cache/<url-hash>.md` yazar, TTL'li (`static.cache_ttl_days`, varsayılan 30). | `--vault` |

Ölçülen: `session-context` bütçenin (200 ms) çok altında; `intent` bütçenin (50 ms)
çok altında — sıcak SQLite cache'in yeniden kullanımı sayesinde.

> **B-025:** `cache-source`'un `PostToolUse`/WebFetch `tool_response` JSON şekli resmi
> dokümandan doğrulanmadı, bu yüzden `cacheBody` bir alan adı tahmin etmek yerine **ham
> baytları** cache'liyor.

---

## 4. Hook kurulumu

Üç ayrı hook ailesi var. Karıştırma.

### 4.1 Claude Code lifecycle hook'ları

| Event | Matcher | Script | Timeout |
|---|---|---|---|
| `SessionStart` | — | `hooks/session-context` | 5 s |
| `UserPromptSubmit` | — | `hooks/user-prompt-intent` | 2 s |
| `SessionEnd` | — | `hooks/session-end-capture` | 10 s |
| `PostToolUse` | `WebFetch` | `hooks/post-tool-cache-source` | 10 s |

Plugin kurulunca `hooks/hooks.json` otomatik yüklenir. Yollar
`"${CLAUDE_PLUGIN_ROOT}"/hooks/...`.

> **Resume tuzağı.** `--continue`/`--resume` ile devam eden bir oturumda `SessionStart`
> yeniden çalışır (beklenen — çıktısı idempotent ve ucuz). Ama **diğer her hook'un
> çıktısı kaydedilmiş transcript'ten tekrar oynatılır**, yeniden çalıştırılmaz. Resume'da
> bayat bir recall hit'i görmek `forge intent`'te bug değil, beklenen davranıştır.
> Bu yüzden hiçbir hook'a zamana duyarlı iş konmaz.

### 4.2 Vault D3 hook'u

```bash
scripts/install_vault_hook.sh <vault-dir> [forge-binary]
```

Idempotent ve **ezmez, reddeder**: forge'a ait olmayan mevcut bir `post-commit` hook'u
varsa kurulum durur.

`<vault>/.git/hooks/post-commit` kurar; `forge capture` çalıştırır.

**Kritik detay:** hook, repo'nun build çıktısını değil **`~/.forge/bin/forge`**'u çağırır.
Mutlak yol `<vault>/.forge/forge-bin`'de pinlidir; `$FORGE_BIN` override eder. Bu bir
**kopyadır** — `pkg/dataset` ya da `cmd/forge/capture.go` değiştiğinde yeniden kur:

```bash
CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge
```

Tasarım gereği hook bir commit'i asla başarısız edemez ve asla bir şey basmaz — bu yüzden
bayat veya bozuk bir binary **sessizdir**. Çiftler görünmeyi bırakırsa:
`<vault>/.forge/capture.log`.

Kaldırma: `rm .git/hooks/post-commit`.

### 4.3 Kod reposu drift hook'ları

```bash
scripts/install_drift_hook.sh <code-repo> <vault-dir> [forge-binary]
```

Üç git-anchored shim kurar: `code-post-commit`, `code-post-merge`, `code-post-checkout`.
Hepsi `forge drift --since-commit <sha>` çalıştırır. Bu da idempotent ve ezmez-reddeder.

- `post-commit` → `HEAD^`'ten beri.
- `post-merge` → **`ORIG_HEAD`**'ten beri (git'in merge-öncesi commit'e koyduğu ref).
  `HEAD^` değil: bir merge tek ebeveynden görünmeyen birçok değişikliği getirebilir.
- `post-checkout` → git üç argüman geçer (`$1` önceki HEAD, `$2` yeni HEAD, `$3` = 1
  branch-seviyesi / 0 dosya-seviyesi checkout); shim yalnızca branch-seviyesinde çalışır.

**Asla dosya kaydında, asla çalışma ağacına karşı.** Commit bir iddiadır; yarı yazılmış
bir dosya değildir.

---

## 5. Skill'ler (slash command'lar)

| Skill | Ne yapar |
|---|---|
| `skills/forge/` | Ana dispatch: soru → recall → pipeline → gate → yaz. Product agent'lara yönlendirir. |
| `skills/forge-init/` | Beş soruyu sorar ve `forge init`'e shell out eder. |
| `skills/forge-check/` | `/forge-check` — haftalık geçişi çalıştırır ve özetler. |
| `skills/forge-stats/` | `/forge-stats` — istatistikleri gösterir. |

Skill'ler **iş mantığı taşımaz**; soruları sorar ve binary'yi çağırır. Bu bilinçli: iş
mantığı test edilebilir Go'da kalsın.

---

## 6. Config referansı

### 6.1 Zincir

```
1. $FORGE_CONFIG                       (en yüksek)
2. <project>/.forge.config.md
3. ~/.forge/forge.config.md
4. config/forge.config.example.md      (gömülü, en düşük)
```

- Eksik opsiyonel katman **atlanır**. Eksik `$FORGE_CONFIG` **hatadır** (kullanıcı onu
  açıkça isimlendirdi).
- **Map'ler anahtar anahtar birleşir. Scalar'lar ve list'ler bütünüyle değişir.**
  Yani bir liste anahtarını (`check.reports`, `static.code_index.languages`) override
  ederken **tamamını** yazman gerekir.
- Dosya markdown, frontmatter'ı YAML. BOM temizlenir, CRLF normalize edilir, `~/` çözülür.

### 6.2 Önemli anahtarlar

```yaml
vault_path: ~/Documents/Base
repo_path: auto

trigger:
  mode: ask                    # ask | auto | manual

recall:
  reuse_threshold: 0.85
  update_threshold: 0.55
  duplicate_threshold: 0.30

freshness_days:
  concept: 365
  howto: 180
  api: 90
  pattern: 365
  pitfall: 365
  incident: 0
  decision: 0
```

```yaml
engines:
  default: host
  api:      { provider: anthropic, model: claude-sonnet-5,
              key_env: ANTHROPIC_API_KEY }
  advisor:  { model: claude-opus-5, mode: critique }
  local:    { enabled: false }
  budget:
    advisor_usd_per_day: 2.00
    api_usd_per_day: 1.00
    on_exhausted: queue
```

```yaml
pipeline:
  intake:     { engine: host }
  recall:     { engine: none }                              # LOCKED
  plan:       { engine: host }
  research:   { engine: api, fallback: host }
  synthesize: { engine: host }
  verify:     { engine: advisor, fallback: local, then: host }
  write:      { engine: none }                              # LOCKED
  link:       { engine: none }
  index:      { engine: none }                              # LOCKED
```

`recall`, `write`, `index` — `none` dışında bir değer yazarsan binary **net bir hatayla
başlamayı reddeder.** Bu kontrol `engine`'in yanı sıra `fallback` ve `then`'e de bakar.

```yaml
static:
  code_index:
    languages: [java, kotlin, python, typescript]
  git_signals: true
  cache_ttl_days: 30
  drift: { ... }
  linkcheck: { ... }
  logback:
    knowledge_map: true
    claude_md_fragment: true
    inline_markers: false      # opt-in

check:
  schedule: "0 9 * * MON"
  churn_days: 90
  duplicate_threshold: 0.40

verify:
  run_code: auto
  duplicate_threshold: 0.40

write:
  language: en
  max_words: 1200
  mermaid: true

telemetry:
  enabled: false
  scope: local
```

### 6.3 Bilerek koda bırakılanlar

Bunlar config'e çıkarılmadı çünkü kullanıcı kararı değil, spesifikasyon sabiti:

- `pkg/vault`'un `excludedPrefixes` / `hubNames`
- `pkg/report/duplicates.go`'nun `specThreshold = 0.85`
- Makefile'daki `$HOME/.forge/bin`

### 6.4 Sekiz paketlenmiş preset

`config/presets/` altında:

| Preset | Grup | Ne için |
|---|---|---|
| `offline.md` | engine | Ağ yok, her şey `none`/`host`. |
| `claude-only.md` | engine | Sadece host tier — API anahtarı gerekmez. **`forge init` varsayılanı.** |
| `byo-api.md` | engine | Kendi Anthropic API anahtarın. |
| `max.md` | engine | Advisor tier dahil, tam bütçe. |
| `java-backend.md` | stack | Java/Spring/Hibernate odaklı. |
| `frontend.md` | stack | React/TS odaklı. |
| `devops.md` | stack | Docker/k8s/CI odaklı. |
| `minimal.md` | stack | Minimum yüzey. |

`forge init --engine-preset` ve `--stack-preset` ile seçilir.

---

## 7. Vault düzeni

```
<vault>/
├── notes/<type>/<slug>.md   concept howto api pattern pitfall incident decision
├── moc/                     Map of Content; moc/weekly/YYYY-WW.md
├── _inbox/                  karantina, confidence: low
├── _archive/
├── profiles/me.md           forge init yazar
├── reports/                 forge check'in çıktısı
├── raw/  sources/           not sözleşmesinin dışında, canlı
├── _index.md                forge index'in çıktısı
└── .forge/                  türetilmiş state, cache, log
```

Yedi not tipi, DESIGN §7'nin beş-dizinli ağacına karşı **B-005**'in kararıdır. Üçü boş
`.gitkeep` kabuğu. **§7'ye uydurmak için budama.**

---

## 8. Sorun giderme

| Belirti | Neden | Çözüm |
|---|---|---|
| `forge: unknown command "logback"` / `"scrub"` | Kök `./forge` binary'si bayat (Phase 5b/6 öncesi). | `make build`, `./dist/forge` kullan. |
| D3 çiftleri artık görünmüyor | `~/.forge/bin/forge` bayat kopya. Hook tasarım gereği sessiz. | `<vault>/.forge/capture.log`'a bak; `CGO_ENABLED=0 go build -o ~/.forge/bin/forge ./cmd/forge`. |
| `forge drift` her şeyi `Skipped` diyor | Silinmiş dosya atıfı, hook yolu değil full sweep. | `--deep` ile çalıştır (B-026), ya da hook yolunda `--apply` (B-028). |
| Config'de yaptığım değişiklik etkisiz | Yanlış katman ya da list-replace semantiği. | `forge config --layers`, sonra `forge config --json`. |
| `engine: pipeline.X: "api" is not allowed` | Kilitli stage'e (`recall`/`write`/`index`) model atanmış. Bu **kasıtlı**. | Config'i düzelt. Override etme yolu **yok**. |
| `forge gate` exit 1 döndü, CI kırmızı | Exit 1 = karantina, **hata değil**. | CI'da 0 ve 1'i başarı say; 2 ve 3'ü hata. |
| `forge check` yavaş / ağ bekliyor | `deadlinks.md` HTTP HEAD atıyor. | `--offline`. |
| Kod indeksleme çalışmıyor, sembol bulunamıyor | Saf-Go lane'de derlenmiş (tree-sitter yok). | `make full` (`CGO_ENABLED=1 -tags codeindex`). |
| `forge recall` iyi bir notu bulmuyor | Bilinen açık defekt **B-008**. | **Eşikleri oynatma.** §3.1 kalibrasyonu yeniden türetilmeli — kendi oturumunu hak ediyor. |
| D2 yakalama hiç çift üretmiyor | **B-024 kapandı** (2026-08-21): `D2Tag` artık `"d2"`, paketlenmiş config'le eşleşiyor. Hâlâ çift yoksa sebep başka: D2 yalnızca gerçek bir advisor çağrısından sonra yazar. | `engines`'te advisor tier'ının yapılandırıldığını ve bütçenin tükenmediğini doğrula. |
| `dataset.capture`'dan `d3`'ü sildim ama yakalama devam ediyor | **B-030**: `d3` için kapı yok — post-commit hook listeye hiç bakmıyor. `d1`/`d5` de okunmuyor; yalnızca `d2` ve `d4` gerçek kapı. | Açık backlog kaydı. Durdurmak için hook'u kaldır: `rm <vault>/.git/hooks/post-commit`. |
| `.forge/code-index-<repo>.json` bekliyordum ama doküman `code-index.json` diyor | **B-027** (kapandı 2026-08-23): davranış doğru (tek isim repolar arası çakışırdı). Kod yorumları ve `forge-codebase-scout` 2026-08-21'de, tasarım dokümanlarındaki sekiz satır 2026-08-23'te düzeltildi. | Kodu esas al: ad her zaman `-<repo>` ekli. |

---

## 9. Geliştirici iş akışı (bu projeyi inşa edenler için)

### 9.1 Faz kuralları

```
0 → 1 → 2 → 2b → 3 → 3b → 4 → 5 → 5b → 6 → [6b]
```

- **Oturum başına bir faz.**
- Faz N merge edilmeden N+1 başlama.
- **2b asla kesilmez.** Zaman biterse kesim sırası: `6b → 5b → advisor tier`.
- Mevcut fazın kapsamı dışında iş çıkarsa → `docs/BACKLOG.md`'ye yaz, inşa etme.

### 9.2 Doküman okuma sırası

0. **`docs/AUDIT.md` §8.4** — bağlayıcı karar kaydı (D-1…D-8). **Önce bunu oku**, çünkü
   tasarım dokümanları bilerek düzenlenmedi: §8.4 bir satırı bayat işaretlediğinde
   doküman hâlâ eskisini söyler ve **takip edilecek olan §8.4'tür.**
1. `docs/ROADMAP.md` — her şeyin üstünde yoğunlaştırılmış index. Hep buradan başla.
2. `docs/KNOWLEDGE-FORGE-STACK.md` (ADR-001) — **her stack sorusunu kazanır.**
3. `docs/KNOWLEDGE-FORGE-DESIGN.md` — master spec.
4. `docs/KNOWLEDGE-FORGE-ADDENDUM.md` — engine tier'ları, dokuz rapor, drift, dataset, config.
5. `docs/CLAUDE-CODE-PROMPT.md` — faz başına yapıştırılmaya hazır prompt.
6. `docs/KNOWLEDGE-FORGE-B2B.md` — **ayrı bir proje**, bu projenin fazı değil (B-021).

### 9.3 Test etme

```bash
make test        # CGO_ENABLED=1 go test ./... ; sonra CGO_ENABLED=0 go build ./...
go test ./pkg/recall -run TestScore -v
go test ./pkg/drift -run TestRollbackSymmetryOnDeletion -v
```

**Her ikisi de gerekli.** Sadece cgo lane'ini test etmek, saf lane'i sessizce kırar.

### 9.4 Fixture kuralları

`testdata/vault/` — 13 not, on iki bilinçli defekt (F1–F12), katalog
`testdata/README.md`.

> **Defektler test yüzeyidir. DÜZELTME.**
> `.git`'i **bilerek yok.** **Asla yerinde `git init` etme** — harness kopyalar ve
> kopyayı init eder.

Bir vault'u mutasyona uğratan her şeyi **önce burada** prova et. Phase 1'in migration'ı
geri alınamazdı ve gerçek vault'un yedeği yoktu.

`examples/vault/` bundan **ayrıdır** — 93 dosya, gerçek vault'tan `forge scrub` ile
üretildi, kapsam `notes/` + `moc/`.

### 9.5 Bilinen tuzaklar

- **20 satır kuralı:** tek bir blok ya da fonksiyonda 20 satırdan fazla kod yazma.
- Vault hook'u binary'nin bir **kopyasını** çağırır — `pkg/dataset` veya
  `cmd/forge/capture.go` değişince yeniden kur.
- `pkg/report`, `pkg/codeindex`'i import **etmemeli** — cgo saf lane'i kırar.
- Modül `knowledge-forge`, dizin `/Users/mimir45/TIL`. Kozmetik uyumsuzluk (B-003/B-004).
  **İstenmeden yeniden adlandırma.**
- `docs/CLAUDE-CODE-PROMPT.md` dokümanların kökte olmasını söyler; `docs/`'talar.
  Prompt metnine uydurmak için dosya taşıma.
