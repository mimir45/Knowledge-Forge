# Knowledge Forge — Dosya Dosya Açıklama

> Fikir için [`01-FIKIR.md`](01-FIKIR.md) · Mimari için [`02-MIMARI.md`](02-MIMARI.md) ·
> Kullanım için [`03-KULLANIM-KILAVUZU.md`](03-KULLANIM-KILAVUZU.md)

**Kapsam:** git tarafından takip edilen **316** dosya (bu doküman dışında), artı
`examples/vault/`'un ayrıca **93** üretilmiş notu → toplam **409**.

**Okuma anahtarı:** her satır `dosya — ne yapar`. Bir dosya bir kararın *nedenini*
taşıyorsa, o neden alıntılanır — çünkü bu repoda doc comment'ler kod kadar spesifikasyon
niteliğindedir.

**İçindekiler**

| # | Bölüm | Dosya |
|---|---|---|
| 1 | Kök | 12 |
| 2 | `.claude-plugin/` | 2 |
| 3 | `.claude/agents/` | 6 |
| 4 | `.github/workflows/` | 2 |
| 5 | `agents/` | 4 |
| 6 | `bin/` | 1 |
| 7 | `cmd/forge/` | 54 |
| 8 | `config/` | 10 |
| 9 | `docs/` | 10 |
| 10 | `evals/` | 7 |
| 11 | `hooks/` | 9 |
| 12 | `pkg/` | 160 |
| 13 | `profiles/` | 2 |
| 14 | `references/` | 6 |
| 15 | `scripts/` | 4 |
| 16 | `skills/` | 4 |
| 17 | `templates/` | 7 |
| 18 | `testdata/` | 16 |
| 19 | `examples/vault/` | 93 (grup) |

---

## 1. Kök (12 dosya)

| Dosya | Açıklama |
|---|---|
| `CLAUDE.md` | **Bu repodaki en önemli tek dosya.** Claude Code'a çalışma talimatı: faz durumu, doküman okuma sırası, agent ekibi, on üç invariant, layout + bütçeler, komut tablosu, "kaydet-düzeltme" tuhaflıkları. Yeni bir oturum bunu okumadan doğru karar veremez. |
| `README.md` | Projenin dış yüzü. **Release-blocking**, sadece dokümantasyon değil: `.goreleaser.yml`'in `archives.files`'ı `README.md`'yi glob-olmayan bir girdi olarak listeliyor — yoksa release build'i kırılır. |
| `LICENSE` | Aynı sebeple release-blocking. |
| `CHANGELOG.md` | Sürüm geçmişi. Phase 6'da eklendi. |
| `CONTRIBUTING.md` | Katkı kuralları: iki build lane, `make test`'in ikisini de çalıştırması, faz disiplini. |
| `Makefile` | Altı hedefli cross-compile matrisi + `build full test bench vet fmt lint dist checksums install-hook clean help`. `install-hook` binary'yi `$HOME/.forge/bin`'e kopyalar — bu yol config'e değil Makefile'a bilerek gömülü. |
| `go.mod` | Modül adı **`knowledge-forge`** (bare path, VCS host prefix'i yok — bilerek ertelendi, **B-004**, açık). Dizin artık `/Users/mimir45/knowledge-forge` (2026-09-01'de kullanıcı tarafından yeniden adlandırıldı, **B-003 kapandı**). |
| `go.sum` | Bağımlılık checksum'ları. |
| `.gitignore` | `dist/`, `.forge/`, build artefaktları. |
| `.golangci.yml` | Lint config. **`errcheck` repo genelinde kapalı** — bilinçli, **B-029** olarak kayıtlı. 2026-08-21'de yeniden ölçüldü: ham **95** bulgu (girdinin "~20" iddiası eksikti), golangci-lint'in varsayılan istisnaları elle uygulanınca ~37 — bunun 10'u üretim kodu. |
| `.goreleaser.yml` | Release paketleme. Altı hedef, arşiv içeriği (README/LICENSE), checksum üretimi. |
| `docs/tr/…` | Bu dört Türkçe doküman (`01-FIKIR`, `02-MIMARI`, `03-KULLANIM-KILAVUZU`, `04-DOSYA-DOSYA`). |

---

## 2. `.claude-plugin/` (2 dosya) — Phase 6'nın paketleme kapanışı

| Dosya | Açıklama |
|---|---|
| `plugin.json` | Claude Code plugin manifest'i. **Phase 0'dan beri açık olan paketleme boşluğunu kapatır:** bundan önce hiçbir şey kök seviyesi `agents/`'ı ya da `hooks/hooks.json`'ı canlı bir `settings.json`'a yüklemiyordu — Claude Code sadece `.claude/agents/`'a bakar. |
| `marketplace.json` | `mimir45/Knowledge-Forge` marketplace girdisi. `claude plugin marketplace add` bunu okur. |

---

## 3. `.claude/agents/` (6 dosya) — **workflow** agent'ları

Bunlar **bu projeyi inşa etmek** içindir. Ürünün dört agent'ı (`agents/`, §5) ile
karıştırma.

| Dosya | Açıklama |
|---|---|
| `finder.md` | *find* fiili. Salt-okunur arama; `file:line` isabetleri raporlar. `/Users/mimir45/Documents/Base` vault'unu da tarar. |
| `executor.md` | *do* fiili. Read/Write/Edit/Bash. Kapsamda kalır, gerçek komut çıktısıyla doğrular. |
| `explainer.md` | *explain* fiili. Salt-okunur — **hiçbir şey yazmaz**, böylece TIL notları `til-writer` skill'inde kalır. |
| `vault-analyst.md` | Salt-okunur vault metrikleri: sayımlar, frontmatter anahtar frekansı, gelen linkler, orphan'lar, near-dupe'lar. |
| `doc-auditor.md` | Tasarım dokümanları arasında **kendilerinin bayrak dikmediği** çelişkileri bulur (Backlog B-001). Çatışmaları raporlar, dokümanı asla düzenlemez. |
| `cross-checker.md` | Başka bir agent'ın sayılarını bağımsız yeniden türetir; strict JSON, iddia başına bir verdict, her biri primary'nin bulgu ID'sine `links`'li. **Primary ile paralel spawn edilir, sonra değil** — cevabı görmüş bir checker ona demir atar. Bu yüzden `vault-analyst` ve `doc-auditor` raporlarını, ID'leri düzyazılarıyla eşleşen bir JSON bloğuyla bitirir: iki koşu mekanik olarak birleşsin diye. |

---

## 4. `.github/workflows/` (2 dosya)

| Dosya | Açıklama |
|---|---|
| `ci.yml` | Build + test + lint + evals. `golangci-lint` kaynaktan derlenir (`install-mode: goinstall`). |
| `release.yml` | Tag'lenmiş release: goreleaser, altı hedef, checksum'lar. Plugin marketplace yolunun beklediği artefaktı üreten şey budur. |

---

## 5. `agents/` (4 dosya) — **ürün** agent'ları (DESIGN §11)

| Dosya | Açıklama |
|---|---|
| `forge-researcher.md` | Araştırma stage'i: kaynak bulur, alıntı toplar. |
| `forge-codebase-scout.md` | Kod tabanını gezer, `code_refs` için aday atıflar üretir. |
| `forge-verifier.md` | Taslağı iddialarına karşı doğrular. |
| `forge-librarian.md` | Notları vault'a yazar/bağlar. **Authored ettiği her commit'e `Forge-Write: true` damgası basar** — basmasaydı `pkg/dataset` kendi çıktısını *insan düzeltmesi* olarak kaydeder ve D3 eğitim verisi kirlenirdi (**B-007**, Phase 4'te kapandı; `pkg/dataset/d3_forge_write_test.go` koruyucuyu iki yönde de pinliyor). |

Phase 6'ya kadar bunlar **canlı, dispatch edilebilir agent değildi** — sadece doğru
spesifikasyondu. `skills/forge/SKILL.md`'nin bunlara dispatch'i bugün açık bir tool
allowlist'iyle generic Agent tool üzerinden doğrulanıyor, canlı agent auto-discovery ile
değil.

---

## 6. `bin/` (1 dosya)

| Dosya | Açıklama |
|---|---|
| `bin/forge` | **Verify-then-exec shim.** Vault'un post-commit hook'u repo'nun build çıktısını değil `~/.forge/bin`'deki bir *kopyayı* çalıştırır ve tasarım gereği o hook bir commit'i asla başarısız edemez, asla bir şey basmaz. "Sessizce bayat ya da değiştirilmiş binary" hiçbir şeyin yüzeye çıkarmadığı tek hata modudur — shim, exec'ten önce sha256 doğrular. |

---

## 7. `cmd/forge/` (54 dosya) — CLI

### 7.1 Giriş ve config

| Dosya | Açıklama |
|---|---|
| `main.go` | 20 komutluk dispatch map'i + kanonik `usage` metni. Bilinmeyen komutta **exit 2**. Başlık doc comment'i `forge engine`'i **sıfır-model-çağrısı kuralının tek istisnası** olarak isimlendirir. |
| `config_cmd.go` | `forge config`. *"Dört katmanlı bir merge kullanıcının kafasında değerlendirebileceği bir şey değil: 'vault path'im neden yanlış' sorusunun cevabı hangi katmanın onu set ettiğidir, ve bu komut olmadan o cevap dört dosya okumayı ve merge kuralını bilmeyi gerektirir."* |
| `config_resolve.go` | Zincir **process başına bir kez** okunur. *"Her subcommand ona ihtiyaç duyuyor ve hiçbiri diğerinden farklı bir cevap görmemeli: `check`'in bir `vault_path` kullandığı, çağırdığı `drift`'in başkasını kullandığı bir koşu iki farklı vault hakkında rapor üretirdi."* |

### 7.2 `forge init`

| Dosya | Açıklama |
|---|---|
| `init.go` | Flag'ler, doğrulama, orkestrasyon. *"Burada hiçbir şey prompt etmiyor. Bu binary'nin içindeki bir TTY döngüsü, farklı bir şapka takmış ikinci bir yazar olurdu: skill artık soran şey olamazdı ve script'lenmiş bir kurulumun içeri girecek yolu kalmazdı."* |
| `init_profile.go` | `<vault>/profiles/me.md` render'ı. `defaultAvoid` DESIGN §9'un listesi birebir ve **bilerek seniority'den bağımsız**: *"hiçbir seviyede kimse notun 'in this article we will' ile açılmasını istemiyor."* |
| `init_write.go` | `configDelta` iki preset overlay'ini, sonra beş cevabı üstüne bindirir. **Sadece delta yazılır:** *"paketlenmiş katman altta kalır ve kullanıcının karar vermediği her şeyi beslemeye devam eder, böylece binary'yi yükseltmek yeni varsayılanları getirir — eskilerinin tam bir kopyasıyla gölgelenmek yerine."* |

### 7.3 Vault temelleri

| Dosya | Açıklama |
|---|---|
| `slug.go` | `forge slug`. `existingSlugs` iddia edilmiş her slug'ı toplar, migration'ın henüz ulaşmadığı notlar için dosya adına düşer — *"yoksa sözleşme-öncesi bir notun adı iki kez dağıtılabilirdi."* |
| `validate.go` | `forge validate`. `applyFix` dosyayı **yalnızca `Fix` gerçekten bir şey değiştirdiyse** yeniden yazar, *"böylece temiz bir koşu her mtime'ı olduğu gibi bırakır ve index cache'i sıcak kalır."* |
| `index.go` | `forge index` / `forge reindex`. `indexTarget` ikisinin de ihtiyaç duyduğu iki ayarı çözer — index dosya adı per-run değil **vault-topoloji** kararıdır: *"onu yeniden adlandıran bir config'e sahip bir vault, reindex `--out`suz çalıştığında ikinci bir `_index.md` almamalı."* |

### 7.4 Recall

| Dosya | Açıklama |
|---|---|
| `recall.go` | `forge recall`. `logAsk` DESIGN §14'ün `ask` event'ini telemetry açıkken kaydeder. `Sources` ve `DurationMS` sıfır kalır: *"forge recall'ın raporlayacak araştırma-süresi ya da atıf-sayısı sinyali yok — bilinen bir sınır, bir eksiklik değil, yukarıdan bir çağıran onu besleyene kadar."* |
| `recall_explain.go` | `--explain` dökümünü **stderr**'e yazar, stdout parse edilebilir JSON kalsın diye. *"İşi, şaşırtıcı bir verdict'i yeniden derleme olmadan debug edilebilir kılmak: hangi kanallar aktifti, her biri ne katkı yaptı ve yeniden normalize eden payda neydi."* |
| `recall_load.go` | `loadDocs` recall'ın vault görünümünü kurar, SQLite cache'i tercih eder. `store.Fresh` (mtime, size) için geçerliyse satır yeniden kullanılır; değilse markdown yeniden parse edilir ve satır yeniden yazılır — *"böylece `forge recall` soğuk ya da kısmi bir cache'i kendi kendine iyileştirir, önce `forge index` çalışmış olmasını gerektirmek yerine."* |

### 7.5 Drift

| Dosya | Açıklama |
|---|---|
| `drift.go` | `forge drift`. `repoList.Set` `--repo name=path`'i parse eder (`strings.Cut(v, "=")`). `driftNotes` atıfları **iki şekilden de** okur: yeni notların yazdığı kanonik `code_refs` bloğu **ve** inline code span'ler — *"migrate edilmiş vault'un elindeki tek şey budur."* Karar mantığı burada değil `pkg/drift`'te. |

### 7.6 `forge check` — dokuz rapor + haftalık geçiş

| Dosya | Açıklama |
|---|---|
| `check.go` | Orkestrasyon. Hedef bir flag değil **sabit**: *"renderer'lar birbirine göreli yolla atıf yapıyor (`see ../moc/codebase.md`), yani `reports/` dışına yazılan bir rapor hiçbir şeye çözülmeyen linkler taşırdı."* |
| `check_collect.go` | Vault'u toplar. **Notlar** (şemanın yargıladığı alt küme; `moc/` ve hub'lar hariç, çünkü *"bir harita için `type:`'ın anlamı yok ve birini saymak geçerlilik paydasını kaydırırdı"*) ile **entry'ler** ayrılır. Koşu sonrası net kontrol: iki sayım tam olarak harita+hub kadar farklı olmalı. Ayrıca `budgetSnapshot`. |
| `check_render.go` | Dokuz raporun renderer'ları, `*checkData` üzerinde metotlar. `values` enum'u `references/schema.yaml`'den okur — *"`coverage.md` onsuz bir yokluğu isimlendiremez: notlarda geçen stack'leri saymak ne hakkında yazıldığını söyler, sadece tam sözlükten çıkarmak ne hakkında yazılmadığını söyler."* |
| `check_codebase.go` | `moc/codebase.md`. ADDENDUM §B.5'in "yüksek churn, gerçek boyut" ifadesini somutlaştırır: *"bir kez dokunulmuş dosya churn etmiyor ve nota sahip olmayan altı satırlık bir accessor bir dokümantasyon boşluğu değil; ikisini de listelemek gerçekten öyle olan bir avuç sembolü gömerdi."* |
| `check_links.go` | `deadlinks.md` için. `cachedOnly` cache'ten cevaplar ve geri kalan her şeyi ulaşılamaz raporlar — *"offline bir koşunun tam olarak öğrendiği şey budur: hiçbir şey."* Alternatif (bilinmeyenleri atlamak) *"paydayı küçültür ve `deadlinks.md`'nin kontrol ettiğinden daha temiz bir vault iddia etmesine sessizce izin verirdi."* |
| `check_asklog.go` | `.forge/log.jsonl`'i okur; `gaps.md` ve hit-rate'i besler. Telemetry kapalıysa log yok. |
| `check_ai_pass.go` | `check.ai_pass`'in üç alt görevi. Her biri `engine.Request` kurar ve **doğrudan `engine.Host{}.Run()`** çağırır — `engine.Resolve`/budget'ı tamamen atlayarak, çünkü *"host tier bir no-I/O talimat basımı, gerçek bir backend çağrısı değil."* **Sadece basar:** buradaki hiçbir şey vault'a yazmaz ya da otomatik uygulamaz. |
| `check_drain.go` | Advisor kuyruğunu boşaltır. *"Bu advisor tier'a gerçek bir HTTP çağrısı, yerel bir kontrol değil."* |
| `check_test.go`, `check_ai_pass_test.go`, `check_asklog_test.go`, `check_drain_test.go` | İlgili testler. |

### 7.7 Engine

| Dosya | Açıklama |
|---|---|
| `engine_cmd.go` | Üç subcommand dispatch'i. *"Bu binary'de gerçek bir model çağrısı yapabilen tek komut ağacı — `main.go`'nun doc comment'i istisnayı isimlendiriyor ve `pkg/engine/select.go` hâlâ recall/write/index'i ondan bağımsız olarak reddediyor."* |
| `engine_select.go` | `selectResult` `--json` şeklidir. **`Engine` kazanan ismi harfi harfine taşır** (`"local"` dahil) *"böylece bir çağıran o durumu düz bir `"api"`'den ayırabilir"*; `Tier` ise `forge engine run`'ın gerçekten dispatch ettiği daraltılmış değer. |
| `engine_run.go` | Gerçek çağrı + bütçe defterine yazım. `queueNote` `pending_advisor: true` damgalar (ADDENDUM §A.4'ün `queue` davranışı). Koşu yine de `none`'a düşer: *"kuyruğa alma bugünün advisor çağrısının ertelendiğini kaydeder, onu satır içinde yeniden denemez."* |
| `engine_record.go` | `engine_trail` damgası. `isLockedStage` **üçüncü** savunma katmanıdır (config load, `engine.Select`, ve bu) — *"record'un doğrulayacak bir config'i yok, sadece kendisine verilen stage adı var."* |
| `engine_run_test.go`, `engine_run_httptest_test.go`, `engine_record_test.go` | Testler; `httptest` olan gerçek HTTP yolunu sahte bir sunucuya karşı sürer. |

### 7.8 Kalite kapıları

| Dosya | Açıklama |
|---|---|
| `gate.go` | `forge gate`. `resolveGateInputs` `--file`/`--rel`'in ikisinin de set olduğunu doğrular ve `--mode`'u parse eder, *"çünkü oradaki bir yazım hatasının sessizce `create`'e düşmesine izin verilmemeli."* Exit kodları: 0 temiz, 1 karantina (**hata değil**), 2 kullanım, 3 iç hata. |
| `verify_code.go` | `forge verify-code`. `readSnippet` kaynak baytları ve dili çözer; `--lang auto` **yalnızca `--file` tanınan bir uzantı isimlendirdiğinde** uygulanır. |

### 7.9 Claude Code hook komutları

| Dosya | Açıklama |
|---|---|
| `session_context.go` | `SessionStart`. `printSessionContext` **kurabildiği her şeyi** yazar — *"index tek başına da yardımcı olur profil eksikse, ve tersi"* — sadece okuma hatalarını loglar. `--max-bytes` index'e ve profile **ayrı ayrı** uygulanır. |
| `intent.go` | `UserPromptSubmit`. `readPrompt` stdin payload'unu decode eder; *"forge intent yalnızca `user_prompt` diye isimlendirdiği tek alana ihtiyaç duyar."* 0.7 üstü isabette `additionalContext` yayar. |
| `session_capture.go` | `SessionEnd`. `maxStubs` bir tetiklemenin yazabileceği stub sayısını sınırlar — planın kendi ifadesi ("küçük bir maks, ör. 3"), *"böylece konuşkan bir oturum `_inbox/`'ı sel altında bırakamaz."* |
| `cache_source.go` | `PostToolUse`/WebFetch. `defaultCacheTTLDays` `Static.CacheTTLDays` unset (sıfır) olduğunda uygulanır — *"config alanının kendi doc comment'iyle eşleşir: sıfır 'unset' demek, 'hemen expire et' değil."* **B-025:** `tool_response` JSON şekli resmi dokümandan doğrulanmadı, bu yüzden `cacheBody` bir alan adı tahmin etmek yerine ham baytları cache'ler. |
| `session_context_test.go`, `intent_test.go`, `session_capture_test.go`, `cache_source_test.go` | Testler. |

### 7.10 Logback (Phase 5b)

| Dosya | Açıklama |
|---|---|
| `logback.go` | Orkestrasyon + flag'ler; `drift`/`check` ile aynı `--vault`/`--repo` şekli, artı `--dry-run` ve `--remove-markers`. |
| `logback_map.go` | `docs/knowledge-map.md` üretimi. |
| `logback_claude.go` | Modül başına `CLAUDE.md` fragment'leri. `logbackSentinelID` **tek bir sabit**, grup başına değil — *"çünkü sentinel bir bloğu id ile **bir dosya içinde** bulur; bir modülün ihtiyacı bir `CLAUDE.md`, bir logback bloğudur."* |
| `logback_markers.go` | Opt-in satır içi `// forge:logback:<symbol>` işaretleri. `markerStyles` sentinel'in Style kümesinden **bilerek dar**: *"`codeindex.Lang` yalnızca Java ve TypeScript parse ediyor, yani buradaki bir Python girdisi işaretleri bu binary'nin asla sembol bulamayacağı bir dile bağlardı. codeindex'in grameri genişlerse bu tablo da genişler."* |
| `logback_test.go` | Dispatch, tam pipeline, idempotent rerun, `--remove-markers` round-trip, per-flag config gating. |

### 7.11 Kalanlar

| Dosya | Açıklama |
|---|---|
| `capture.go` | `forge capture` — D3 hasadı; vault'un post-commit hook'unun çağırdığı şey. |
| `scrub.go` | `forge scrub`. Usage'ı sözleşmeyi net söyler: *"Fails closed: bir not scrub'lanıp `references/schema.yaml`'e karşı yeniden doğrulanamıyorsa `--dst`'ye hiçbir şey yazılmaz — temiz scrub'lanmış notlar bile."* Exit 0/2/3. |
| `stats.go` | `forge stats`. `estimatedMinutesSavedPerHit` **bilerek kaba**: *"bu projede bir 'zaten sorduk ve cevapladık' yeniden-açıklamasının birine gerçekten ne kadara mal olduğuna dair hiçbir ölçüm yok, bu yüzden bu isabetleri zamana yalnızca bir büyüklük-mertebesi sinyali olarak çevirir, asla gerçek bir metrik olarak değil."* Hem doc comment hem basılan satır ona "yaklaşık" der. |
| `drift_test.go`, `recall_test.go`, `stats_test.go`, `e2e_test.go` | Testler; `e2e_test.go` uçtan uca CLI koşuları. |

---

## 8. `config/` (10 dosya)

| Dosya | Açıklama |
|---|---|
| `embed.go` | `//go:embed` — örnek config ve sekiz preset'i binary'ye gömer. Bu, *"bir tarball'dan `forge` çalıştıran bir yabancının yine de eksiksiz bir config almasını"* sağlayan mekanizma. |
| `forge.config.example.md` | **Paketlenmiş taban katman** (zincirin 4. katmanı). Kullanıcılar bunu düzenlemez; `forge init` `~/.forge/forge.config.md`'ye yalnızca delta yazar. `config/forge.config.md` diye bir dosya **yoktur ve olmamalı** — o bir paketlenmiş şablon adı, bir yazma hedefi değil. |
| `presets/offline.md` | Engine preset: ağ yok. |
| `presets/claude-only.md` | Engine preset: yalnızca host tier, API anahtarı gerekmez. **`forge init` varsayılanı.** |
| `presets/byo-api.md` | Engine preset: kendi Anthropic anahtarın. |
| `presets/max.md` | Engine preset: advisor tier dahil, tam bütçe. |
| `presets/java-backend.md` | Stack preset: Java/Spring/Hibernate. |
| `presets/frontend.md` | Stack preset: React/TypeScript. |
| `presets/devops.md` | Stack preset: Docker/k8s/CI. |
| `presets/minimal.md` | Stack preset: minimum yüzey. |

---

## 9. `docs/` (10 dosya)

**Okuma sırası kritik.** Bkz. `03-KULLANIM-KILAVUZU.md` §9.2.

| Dosya | Açıklama |
|---|---|
| `AUDIT.md` | **Önce bunu oku.** §8 on üç çelişkiyi listeler (dokümanların kendilerinin bayrak dikmediği); sekizi precedence kuralıyla çözülür. **§8.4 bağlayıcı bir karar kaydıdır (D-1…D-8)** ve precedence'ın çözemediği altısını kapatır. Tasarım dokümanları **bilerek düzenlenmedi**: §8.4 bir satırı bayat işaretlediğinde doküman hâlâ eskisini söyler ve **takip edilecek olan §8.4'tür.** |
| `ROADMAP.md` | Her şeyin üstünde yoğunlaştırılmış index. Her zaman buradan başla. |
| `KNOWLEDGE-FORGE-STACK.md` | **ADR-001. Her stack sorusunu kazanır.** ADDENDUM §B'yi (Python'u belirtiyordu — dokümanın kendisi "bu yanlıştı" diyor) ve B2B §8'i (Spring Boot varsayıyordu — artık açık bir karar, ADR-002) geçersiz kılar. |
| `KNOWLEDGE-FORGE-DESIGN.md` | Master spec: şema, pipeline, kapılar, vault topolojisi, subagent'lar. Rev-2 notu gereği her `scripts/*.py` referansı bir `forge` subcommand'ı olarak okunur. |
| `KNOWLEDGE-FORGE-ADDENDUM.md` | Engine tier'ları (§A), AI-siz yetenek sınırı ve raporlar (§B), drift (§B.6), haftalık checker (§C), dataset'ler (§D), tam config YAML + preset'ler (§E). |
| `CLAUDE-CODE-PROMPT.md` | Gerçek yürütme mekanizması: faz başına yapıştırılmaya hazır bir prompt. *(Dokümanların repo kökünde olmasını söyler; `docs/`'talar. Prompt metnine uydurmak için dosya taşıma.)* |
| `BACKLOG.md` | Açık maddeler. Mevcut fazın kapsamı dışında iş çıkarsa **inşa etme, buraya yaz.** |
| `KNOWLEDGE-FORGE-B2B.md` | **Ayrı bir proje**, bu projenin bir fazı değil (**B-021**). Repoda yalnızca referans/tarih için. |
| `adr/0001-lexical-recall-vs-embeddings.md` | DESIGN §8'den: neden leksik recall, neden embedding yok. |
| `adr/0002-go-for-static-core.md` | STACK §1'den: neden statik çekirdek Go. |

STACK §6 için üçüncü bir ADR **yok** — B2B ayrı bir proje olduğundan (B-021).

---

## 10. `evals/` (7 dosya)

| Dosya | Açıklama |
|---|---|
| `README.md` | Eval iskelesinin ne ölçtüğü. |
| `run.sh` | Eval koşucusu; CI'ın lint/evals adımı bunu çağırır. |
| `fixtures/vault/notes/concept/kafka-consumer-group-rebalancing.md` | Eval fixture notu. |
| `fixtures/vault/notes/howto/configure-kafka-consumer-timeouts.md` | İkinci eval fixture notu — bu ikisi kasten ilişkili, recall'ın `reuse`/`update` ayrımını sürmek için. |
| `fixtures/vault/moc/.gitkeep`, `_inbox/.gitkeep`, `_archive/.gitkeep`, `profiles/.gitkeep` | Fixture'ın gerçek vault topolojisini taşıması için boş dizin kabukları. |

---

## 11. `hooks/` (9 dosya)

Hepsi POSIX `sh`. Shim'lerin tek işi forge binary'sini bulmak ve stdout/stderr'ı geçirmek
— **iş mantığı Go'da.** Hepsi fail-silent, hepsi exit 0.

| Dosya | Açıklama |
|---|---|
| `hooks.json` | Dört binding, hepsi `${CLAUDE_PLUGIN_ROOT}`-göreli: `SessionStart`→`session-context` (5 s), `UserPromptSubmit`→`user-prompt-intent` (2 s), `SessionEnd`→`session-end-capture` (10 s), `PostToolUse` matcher `"WebFetch"`→`post-tool-cache-source` (10 s). |
| `session-context` | `forge session-context` çalıştırır; trim'lenmiş `_index.md` + geliştirici profilini stdout'a basar — o çıktı taze oturumun context'ine düşer. |
| `user-prompt-intent` | `forge intent` çalıştırır; stdin JSON'ının `user_prompt` alanını okur, ucuz model-siz recall kontrolü yapar ve **yalnızca en iyi isabet 0.7 üstündeyse** `{"additionalContext": "...", "continue": true}` basar. |
| `session-end-capture` | `forge session-capture` çalıştırır; stdin JSON'ını (`session_id`, `transcript_path`, `reason`) okur, transcript'in assistant-rol metnini "we established/decided/concluded/agreed that…" cümleleri için regex tarar, **en fazla 3** stub yazar. |
| `post-tool-cache-source` | `forge cache-source` çalıştırır; `tool_name` `"WebFetch"` olduğunda `<vault>/.forge/cache/<url-hash>.md`'yi TTL header'ıyla yazar. `hooks.json`'ın matcher'ı zaten kapsamı daraltıyor — shim'in kontrolü ikinci savunma. |
| `vault-post-commit` | **VAULT** reposunun `.git/hooks/post-commit`'i. Her commit'ten sonra `forge capture`'a o commit'in son yedi gün içinde forge'un ürettiği bir notu düzenleyip düzenlemediğini sorar; öyleyse (üretilmiş, insan-tercihli) çifti D3'e ekler. |
| `code-post-commit` | **KOD** reposunun post-commit'i. `HEAD^`'ten beri drift. |
| `code-post-merge` | Kod reposunun post-merge'ü. **`ORIG_HEAD`**'ten beri — *"git'in merge-öncesi commit'e koyduğu ref; `HEAD^` değil, doğru since-SHA budur: bir merge tek ebeveynden görünmeyen birçok değişiklik getirebilir."* |
| `code-post-checkout` | Kod reposunun post-checkout'u. Git üç argüman geçer (`$1` önceki HEAD, `$2` yeni HEAD, `$3`=1 branch-seviyesi / 0 dosya-seviyesi); shim yalnızca branch-seviyesinde çalışır. |

---

## 12. `pkg/` (160 dosya, 18 paket)

### 12.1 `pkg/vault/` (18) — markdown gerçeğin tek kaynağı

| Dosya | Açıklama |
|---|---|
| `note.go` | `Note` = disktekinde bir markdown dosyası, *"doğrulanacak ve indekslenecek kadar parse edilmiş."* |
| `frontmatter.go` | YAML frontmatter parse'ı. `ErrNoFrontmatter` = hiç `---` bloğu yok. *"Gerçek vault'ta on beş not bu durumda, yani bu beklenen bir koşul, bir parse hatası değil."* BOM ve CRLF burada normalize edilir. |
| `schema.go` | `Field` = `references/schema.yaml`'deki `fields:` altındaki bir girdi. *"Şemanın ifade edebildiği her kısıtın burada bir evi var; nil pointer 'kısıtlanmamış' demek."* |
| `validate.go` | `Issue` = bir doğrulama hatası. *"`Code` stabil ve makine-greplenebilir; `Msg` eyleme geçirilebilir yarısı. `Fixable` `forge validate --fix`'in mekanik onarabileceklerini işaretler — asla bir yargı çağrısı, yalnızca tarih backfill'i, anahtar sırası, case ve alias yeniden yazımı."* |
| `fix.go` | Mekanik onarımlar. `ForgeVersion` anahtarı eksik olan notlara damgalanır. |
| `links.go` | Wikilink çıkarımı. `StripCode` her fenced blok ve inline code span'i tek boşlukla değiştirir. **Export edilmiş** çünkü *"aynı 'kod düzyazı değildir' filtresine ihtiyaç duyan paket dışı çağıranlar — `pkg/qualitygate`'in anti-slop yasaklı-ifade taraması ilki — Wikilinks'in kendi ilk adımını yeniden yazmak zorunda kalmasın."* |
| `slug.go` | `Slug` bir başlığı kanonik kebab-case'e çevirir. **Deterministik:** *"aynı girdi hep aynı çıktıyı verir; saat, map iterasyonu ya da rastgelelik yok."* Vault-genelinde teklik için `SlugUnique`. |
| `write.go` | Frontmatter damgalama. `ErrNoFM` bir çağıran damgalayacak frontmatter bloğu olmayan bir notu damgalamaya çalıştığında döner — *"drift bunlarla vahşi doğada karşılaşıyor (fixture'da üç tane var) ve bir header uydurmak yerine atlamalı."* |
| `quarantine.go` | `WriteToInbox`. Kendisi *"CREATE'i UPDATE'ten bilmez"* — çağıran söyler. |
| `note_test.go` … `write_test.go` (8 test) + `bench_test.go` | Testler ve benchmark. |

### 12.2 `pkg/recall/` (9) — deterministik skorlama

| Dosya | Açıklama |
|---|---|
| `doc.go` | `Doc` = recall'ın gördüğü bir not: frontmatter + gövdeye tembel bir handle. *"Gövde bir func, çünkü DESIGN §8 adım 3 yalnızca ilk 20 adayı okuyor — diğer kanallar önce sıralar ve çoğu not hiç açılmaz."* |
| `normalize.go` | Tokenizasyon + stopword'ler. Stopword'ler *"skill'i tetikleyen ifadelerin iskelesi — 'how does X work', 'what is the difference between X and Y', 'best practices for X'. Bırakılsalardı hiçbir başlığın eşleyemeyeceği 3–5 token olurlar ve mükemmel bir isabetin tavanı üçte bir düşerdi."* |
| `score.go` | Kanal ağırlıkları, *"DESIGN §8'in blend'i birebir. Oranları tasarımın sabitlediği şey; bölündükleri payda §2.5'in karar verdiği şey."* — yani **aktif kanallar üzerinden ağırlıklı ortalama**, literal ağırlıklı toplam değil. Başlık ölçüsü **F₂, Dice değil**. |
| `rank.go` | `BodyPassSize` DESIGN §8 adım 3'ün "ilk 20 dosya"sı. *"Üç frontmatter kanalı önce sıralar ve yalnızca liderler açılır; gövde geçişini birkaç yüz kilobayta sığmayan bir vault'ta ucuz tutan şey budur."* |
| `freshness.go` | `parseDate` hem şemanın `YYYY-MM-DD`'sini hem *"Obsidian eklentilerinin bazen bunun yerine yazdığı RFC3339 timestamp'leri"* kabul eder. |
| `result.go` | Verdict zarfı — eşik ağacı burada, downstream'de değil. |
| `rank_test.go`, `result_test.go`, `score_test.go` | Testler. |

> **Açık defekt B-008.** Reçete edilen IDF ağırlıklandırma ship edildi ve isimlendirilen iki
> vakanın hiçbirini düzeltmedi: *bir sorunun anlamını taşıyan terimler, hiçbir not onları
> taşımadığında paydadan filtreleniyor.* Sonraki deneme **§3.1 kalibrasyonunun tamamını
> yeniden türetmeye** sahiplik ediyor. **Buna eşikleri oynatarak cevap verme.**

### 12.3 `pkg/similarity/` (3) — MinHash + LSH

| Dosya | Açıklama |
|---|---|
| `minhash.go` | `Hashes` imza uzunluğu. *"128, Jaccard tahmininin standart hatasını 1/√128 ≈ %9'a koyar; adayları bir insanın sonra okuyacağı şekilde sıralayan bir rapor için bu yeterli."* |
| `lsh.go` | Banding. Bir düzeltmenin kaydı: eski parametreler P(0.40) = 0.56 veriyordu ve *"her biri boş rapor döndürürken Estimate çiftin duplicate olduğunu kabul ediyordu."* Düzeltmenin bedeli aday hacmi — s=0.10'da 64×2 tüm çiftlerin yaklaşık yarısını aday gösteriyor, *"ki bu gerçek vault'ta 1142 imza karşılaştırması ve bir vault ölçülecek kadar büyüyene dek optimize etmeye değmez."* |
| `similarity_test.go` | Testler. |

**Embedding yok.** Elle yazılmış, deterministik.

### 12.4 `pkg/graph/` (4)

| Dosya | Açıklama |
|---|---|
| `graph.go` | `Node` = bir notun graftaki konumu; kenarlar wikilink'lerden. |
| `components.go` | `Component` = bağlantılı bir not grubu. `graph-health.md` ve `orphans.md`'yi besler. |
| `graph_test.go`, `components_test.go` | Testler. |

### 12.5 `pkg/codeindex/` (8) — **tek cgo paketi**

| Dosya | Açıklama |
|---|---|
| `index.go` | `Extractor` ve `ErrUnavailable` (*"cgo'suz derlendiğinde `Parse` bunu döner"*). Doc comment'i **cache-format/serialize-edilmiş-şekil versiyonlamayı** açıkça kapsar, yalnızca extraction-logic versiyonlamayı değil — kendi "ilk release edilmiş binary'den önce inmeli" metnine göre (**B-013**, Phase 6'da kapandı). |
| `parse_cgo.go` | go-tree-sitter yolu. `Available()` burada `true`. |
| `parse_nocgo.go` | Saf-Go yolu. `Available()` `false`, *"böylece çağıranlar nil sonuç yerine net bir mesajla degrade olabilir."* |
| `build.go` | İndeksi kurar; sonuç *"tree state'in saf bir fonksiyonu — drift'in verdict'lerinin bağlı olduğu aynı özellik."* |
| `catfile.go` | `git cat-file --batch` sürücüsü. `drainBlobs` cevapları **istek sırasında** yürür — *"git girdi satırı başına bir tane garanti ediyor, yani anahtar istenen yol, cevaptaki hiçbir şey değil."* |
| `deps.go` | `Deps` deklare edilmiş bağımlılık versiyonlarını repo'nun sahip olduğu build manifest'lerinden okur. *"Drift'in beşinci verdict'i — 'deklare edilmiş dep versiyonu yükseldi, not eski davranışı anlatıyor olabilir' — bu map'lerden ikisinin karşılaştırması, yani birim map, tek tek dosya değil."* |
| `store.go` | `Save` **çağıranın verdiği yola** yazar; bu paket kendi dosya adını belirlemez (`pkg/drift` repo başına bir dosya yazar). *"`.forge/` altındaki diğer her dosya gibi türetilmiş bir cache: silmek bir yeniden kurulum maliyeti ve başka hiçbir şey."* Doc comment'i 2026-08-21'de düzeltildi — eskiden tekil `.forge/code-index.json` iddia ediyordu; bkz. **B-027**. |
| `parse_test.go` | Testler. |

### 12.6 `pkg/coderef/` (6) — atıf çıkarımı ve çözümü

| Dosya | Açıklama |
|---|---|
| `ref.go` | `Kind` *"extractor'ın bulduğu şeyden ne kadar emin olduğunu kaydeder, ki çözümün sonra ihtiyacı olur: çıplak bir CamelCase token bir sembol *adayıdır* ve sırf adını taşıyan bir dosya yok diye broken raporlanmamalı."* **Kritik:** kanonik `code_refs:` biçimi (`repo:path#Symbol`) `KindPath` + `Symbol` set olarak parse olur. |
| `extract.go` | `sourceExt` bir atıfın isimlendirebileceği uzantı kümesi. **Bilerek dar:** *"vault, kod referansı değil konfigürasyon muhabbeti olan `.md`, `.yml` ve `.json` span'leriyle dolu ve her biri hiçbir şeye çözülüp NF-4'ün dürüstçe ölçmeye çalıştığı unresolved sayısını şişirirdi."* |
| `resolve.go` | `Repo` = vault'un atıf yaptığı bir kod reposu. `Name` kanonik bir ref'in söylediği şey; `Files` forward-slash'li repo-göreli yollar. |
| `scan.go` | `runGit` tek bir git subcommand'ı kök altında çalıştırır. *"Aşağıdaki her fonksiyon bunun üzerinden shell out ediyor — `pkg/gitsig` aynı CLI-değil-go-git seçimini yapıyor (B-009) — böylece `exec.Command` boilerplate'i dörde değil bir yere yaşıyor."* |
| `extract_test.go`, `resolve_test.go` | Testler. |

### 12.7 `pkg/gitsig/` (4)

| Dosya | Açıklama |
|---|---|
| `log.go` | `Commit` = bir commit ve dokunduğu dosyalar. |
| `stats.go` | `Stats` = bir commit aralığından türetilen dosya-başına sinyaller. |
| `rank.go` | `FileCount` = bir dosya ve kaç commit'in ona dokunduğu. |
| `gitsig_test.go` | Testler. |

**Bilinçli sapma B-009:** go-git yerine `git` CLI'ına shell out eder.

### 12.8 `pkg/drift/` (9) — **kilit paket**

| Dosya | Açıklama |
|---|---|
| `drift.go` | `Verdict` = bir atıfın sonucu (`OK`/`Repaired`/`Suspect`/`Broken`/`Skipped`). Sözleşme tipleri. |
| `check.go` | `checkPath` ADDENDUM §B.6'nın merdivenini **addendum'un belirttiği sırayla** yürür: dosya gitti → sembol gitti → satır kaydı → gövde değişti. *"Sıra önemli: silinmiş bir dosya aksi halde kaymış satır olarak raporlanırdı."* |
| `gitsource.go` | `Repo` = drift'in bakmasına izin verilen kod reposu. `Source` arayüzü (`At`/`RevBefore`/`Head`/`Find`/`ResolveAt`) `pkg/drift`'i saf Go tutan şeydir — tree-sitter buranın arkasında kalır. |
| `gitindex.go` | `build` kalıcı indeksi tercih eder ve onu ileri doğru yamalar, *"böylece hook yolundaki tek tree-sitter işi commit'in dokunduğu bir avuç dosya. HEAD'in bir commit gerisinde bir cache **normal** durumdur — hook post-commit ateşliyor — ve 'HEAD'deki sembol tablosu' ifadesini yaklaşık değil gerçek yapan şey yamalamadır."* Cache adı repo başına: `.forge/code-index-<repo>.json`. `persist`'in doc comment'i 2026-08-21'den beri sapmanın **neden** gerekli olduğunu açıklıyor: `--repo` tekrarlanabilir olduğu için tek paylaşılan ad, ikinci repo'nun indeksinin birincisininkini ezmesine yol açardı (**B-027**). |
| `apply.go` | `Result` = bir notun confidence hareketi, CLI çıktısı ve `drift.md` için. `--apply` olmadan hiçbir şey taşınmaz. |
| `demotions.go` | Demote geçmişi. Sadece demote-öncesi confidence saklanır — *"en fazla bir notun confidence'ının ne olduğunun hafızasına mal olur, yanlış bir cevaba değil."* Not gövdesinde asla. |
| `check_test.go`, `gitindex_test.go`, `rollback_test.go` | Testler. `rollback_test.go` `TestRollbackSymmetryOnDeletion`'ı içerir — **B-028**'in kapanış kanıtı: eşleşmeyen bir hook-yolu ıskası **hiç bulgu üretmez, asla `Skipped` değil**, böylece alakasız bir sonraki commit hâlâ bozuk bir notu `high`'a geri çeviremez. |

### 12.9 `pkg/linkcheck/` (4)

| Dosya | Açıklama |
|---|---|
| `linkcheck.go` | `Verdict` = bir kontrolün vardığı sonuç. |
| `probe.go` | HTTP HEAD. `UserAgent` checker'ı tanımlar — *"boş ya da varsayılan bir Go user agent, birkaç doküman host'unun 403 ile cevapladığı şeydir, ki bu ölü link olarak okunurdu."* |
| `cache.go` | `CacheFile` cache'in `New`'e verilen dizin içindeki adı. Rate-limit + kalıcılık. |
| `linkcheck_test.go` | Testler. |

### 12.10 `pkg/report/` (20) — render katmanı

**Kural: `pkg/report`, `pkg/codeindex`'i import etmemeli** — cgo saf lane'i kırardı.

| Dosya | Açıklama |
|---|---|
| `report.go` | `Report` = `<vault>/reports/`'a giden bir render edilmiş dosya. |
| `index.go` | `Entry` = index'in ihtiyaç duyduğu haliyle bir not. `_index.md` render'ı. |
| `coverage.go` | `coverage.md`. Tam stack sözlüğü olmadan *"rapor bir yokluğu isimlendiremez."* |
| `staleness.go` | `staleness.md`. |
| `duplicates.go` | `duplicates.md`. Spec eşiği `specThreshold = 0.85` **bilerek kodda**, config'de değil. |
| `orphans` / `graph.go` | `OrphansInput` = `orphans.md`'nin render kaynağı; graph-health de burada. |
| `gaps.go` | `Ask` = sorulmuş bir konu, kaç kez ve şu an bir notun cevaplayıp cevaplamadığı. |
| `deadlinks.go` | `Citation` kontrol edilen bir URL'i onu alıntılayan notlara geri bağlar. *"Bir URL sık sık birkaç not tarafından alıntılanır ve rapor ancak onları isimlendirirse eyleme geçirilebilir."* |
| `churn.go` | `ChurnInput`. İstatistikler **vault'un** geçmişi üzerinedir, bir kod reposununki değil: *"§B.4'ün `churn.md`'si 'hangi notlar sürekli yeniden yazılıyor' sorusunu cevaplar. Kod churn'ü `moc/codebase.md`'de yaşar ve ikisi karıştırılmamalı — aynı ölçüm, farklı repolar üzerinde, ve farklı şeyler ifade ediyorlar."* |
| `drift.go` | `DriftInput` = `drift.md`'nin render kaynağı: kod referansı başına bir bulgu, `forge drift --deep`'in ürettiği gibi. |
| `codebase.go` | `CodeGroup` = `moc/codebase.md`'nin ihtiyaç duyduğu haliyle bir modül/paket. |
| `knowledgemap.go` | `RenderKnowledgeMap` (Phase 5b). `DependsOn` `CodeGroup`'ta var ama *"bu kod tabanında hiçbir şey onu doldurmuyor, bu yüzden hep boş okunacak bir sütun render etmek yerine burada da atlanıyor."* |
| `cost.go` | `CostInput` — ADDENDUM'un "stage başına tier başına token/$" tablosu. *"`pkg/report` config-free kalır, buradaki her diğer raporla eşleşerek: `cmd/forge/check_collect.go` SQLite budget tablosunu ve `cfg.Pipeline`/`cfg.Engines`'i bu primitive'lere indirger."* |
| `weekly.go` | `VaultStats` = bir haftanın manşet sayıları, snapshot'lanmış ki sonraki koşu delta gösterebilsin. `HitRate` **kümülatif**, haftaya filtrelenmemiş — *"DESIGN §14 event'leri timestamp taşıyor ama `loadAskLog`'un downstream'inde hiçbir şey henüz ona göre pencerelemiyor. Bu bilinen bir sadeleştirme, bir bug değil."* |
| `weekly_store.go` | `.forge/weekly-stats.json` kalıcılığı. `WeekKey` ISO haftasını `"YYYY-Www"` olarak zero-padded formatlar *"böylece düz string karşılaştırması haftaları doğru sıralar — yıl sınırı boyunca dahil, ki orada ISO haftasını anahtarlayan `ISOWeek`'in kendi döndürdüğü yıldır, `t.Year()` değil."* |
| `index_test.go`, `reports_test.go`, `knowledgemap_test.go`, `weekly_test.go`, `weekly_store_test.go` | Testler. |

### 12.11 `pkg/store/` (5) — SQLite türetilmiş cache

| Dosya | Açıklama |
|---|---|
| `store.go` | `Store` cache veritabanını sarar. `modernc.org/sqlite` — saf Go, cgo yok. Şema burada. |
| `read.go` | `attach` bir çocuk tablonun değerlerini zaten yüklenmiş satırların üstüne katlar. |
| `budget.go` | **Tek istisna.** `budgetSchemaSQL` `schemaSQL`'den ayrı yaşar *"böylece `Reset()`'in asla dokunmaması gereken tek tablo, `store.go`'nun listesine gömülü bir satır yerine tek-dosyalık bir diff olarak kalır (AUDIT §8.4 D-8)."* Yani **budget `forge reindex`'ten sağ çıkar.** |
| `store_test.go`, `budget_test.go` | Testler. |

### 12.12 `pkg/config/` (7) — dört katmanlı zincir

| Dosya | Açıklama |
|---|---|
| `types.go` | `Config` birleştirilmiş zincir. *"Her alan birleşim şeması; Go anlamında hiçbir şey opsiyonel değil, çünkü paketlenmiş katman her zaman bir değer sağlıyor."* |
| `load.go` | Zincir yürütme. `PackagedName` gömülü taban katmanın hatalarda ve `Layers`'ta adı — *"diskte bir yol değil: örnek binary'ye derleniyor, böylece bir tarball'dan `forge` çalıştıran bir yabancı yine de eksiksiz bir config alıyor (D-2 — paketlenmiş katman kullanıcılar tarafından asla düzenlenmiyor, yani kaybedebilecekleri bir dosya olması için bir sebep yok)."* |
| `merge.go` | Anahtar-anahtar map merge; scalar/list bütünüyle replace. Örnek gerekçe: *"bir kullanıcı katmanı `verify: {fallback: host}` set ederken §E'nin `engine: advisor`'ını korumalı."* |
| `preset.go` | `Preset` bir preset'i **merge overlay'i olarak** döner. *"Bir `Config` değil: bir preset bir avuç anahtar set eder ve geri kalanını miras alır, yani onu struct'a decode etmek set edilmemiş her alanı, sonra paketlenmiş katmanı ezen bir zero value'ya çevirirdi."* |
| `validate.go` | Kilitli stage doğrulaması. *"…index, `forge reindex`'i cache'i yeniden kuramaz hale getirir. Aksini söyleyen bir config'te kural **net bir hatayla başlamayı reddetmektir** — asla sessizce override etme, çünkü `engine: host` yazıp yine de `none` alan bir kullanıcıya yalan söylenmiştir."* |
| `chain_test.go`, `helpers_test.go` | Testler. |

### 12.13 `pkg/engine/` (17) — dört tier

| Dosya | Açıklama |
|---|---|
| `engine.go` | `Tier` = bu paketin ship ettiği dört implementasyondan biri. Ortak `Engine` arayüzü. |
| `none.go` | T0. *"Sıfır I/O, her zaman reddeder. Gerçek bir `Engine` değeri olarak var (nil özel-durumu yerine) ki bir `Engine` arayüzü tutan çağıranlar bir stage'in statik olduğunu öğrenmek için asla type switch'e ihtiyaç duymasın."* |
| `host.go` | T1. *"Go binary'sinin kendisinin çağıramayacağı tier için seam: model bu process'te değil Claude Code oturumunun içinde çalışıyor. `Run` hiç I/O yapmaz — skill'in isteği kendisi yürütmesi için gereken her şeyi geri verir, ve skill sonucu `forge engine record` üzerinden geri bildirir, böylece `engine_trail` yine damgalanır."* |
| `api.go` | T2. **Boş olmayan bir `BaseURL`, Provider'ı `"ollama"` olan bir API değerine çözülür** — yerel barındırılan bir model sunucusunun konuştuğu şekil. Yani `local` beşinci bir Engine değil, farklı bir base_url altında `api.go`. |
| `api_provider.go` | `payload` provider'a özgü istek gövdesini kurar. *"Bu, `api.go`'nun provider başına gerçekten farklılaşan tek parçası — cevap zarfı farklılaşmıyor."* |
| `advisor.go` | T3. *"Critique-only. Bir notu asla yeniden yazmaz — DESIGN'ın sözleşmesi ihtilaflı iddialar, eksik olan, bir confidence verdict'i ve minimal bir patch; ve bu tip, o sözleşmeyi bir prompt string'inde ifade edilmiş bir umut yerine yapısal tutan şey."* |
| `select.go` | `Resolve`/`Select`/`Exhausted`/`tierOf`/`checkLocked`/`isLocked`/`chain`. `LockedStages` `pkg/config`'in listesini yeniden export eder — *"defense in depth, onun etrafından dolanan bir kısayol değil: `config.Load` zaten aynı ihlalde `validateLockedStages` üzerinden başlamayı reddediyor."* `checkLocked` `Engine`'in yanı sıra `Fallback` ve `Then`'e de bakar: *"`pipeline.write.fallback`'in arkasına saklanan bir tamper da burada yakalanmalı, yoksa bu katman dekoratif."* `chain`: *"set edilmemiş bir stage `none`'a-kilitli iddiası değil, sessizliktir, ve onu dolduran `cfg.Engines.Default`'tur."* |
| `availability.go` | `available` bir adayın şu an çalışıp çalışamayacağını **ve nedenini** raporlar — *"reason string'i `select.go`'nun `forge engine select --json`'ın 'reason'ı için geri verdiği şey."* |
| `budget.go` | `Ledger` bu paketin ihtiyaç duyduğu budget store'u. *"Yapısal olarak `*store.Store` ile eşleşiyor, böylece `pkg/engine` `pkg/store`'u asla import etmiyor ve leaf kalıyor. `Spend`/`Remaining` enjekte edilmiş bir clock alıyor, böylece bir test `time.Now()` ile yarışmak yerine bir günü pinleyebiliyor."* |
| `trail.go` | `engine_trail` damgası. Doc comment'i şema geçmişini kaydeder: dokuz gerçek stage adı, ve *"şemanın eski `critique` alternatifi düşürüldü — o hiç gerçek bir stage değildi, sadece advisor engine'inin çalışma modu, `engines.advisor.mode`"*. |
| `none_test.go`, `host_test.go`, `api_test.go`, `advisor_test.go`, `select_test.go`, `trail_test.go`, `trail_entry_test.go` | Testler. |

### 12.14 `pkg/qualitygate/` (20) — yedi kapı

| Dosya | Açıklama |
|---|---|
| `gate.go` | `Run`/`Report` orkestrasyonu. `Remedy` *"bir başarısız kapının önerdiği şey. Tavsiye niteliğinde — `Run` ona asla göre davranmaz, `quarantine.go` ve `forge gate` CLI'ı davranır — çünkü DESIGN §12 her kapıya farklı bir hata cevabı veriyor ve onları tek bir Fail/Pass bitine indirmek, skill'in doğru davranmak için ihtiyaç duyduğu ayrımın tam olarak kendisini kaybederdi."* `blocksWrite` hangi remedy'lerin bloklayacağını belirler. |
| `schema.go` | 1. kapı. *"`vault.Validate` etrafında ince bir sarmalayıcı — `forge validate`'in çalıştırdığı aynı kontrol, böylece `forge validate --all`'ı geçemeyecek bir taslak diske ulaşmadan önce burada da geçemiyor. Herhangi bir issue kapıyı düşürür: DESIGN §12 eksik anahtarla bilinmeyen anahtar arasında ayrım yapmıyor, ikisi de yazmayı bloke ediyor."* |
| `citation.go` | 2. kapı. `checkSourcesArity` zaten aynı olguda `RetryOnce` ile bloke ediyor; *"bu kapı onu ikinci kez `MarkUnverified` ile raporluyor, böylece citation hatalarını yumuşak (yayınla ama bayrakla) sayan bir çağıran o ayrıma göre davranabiliyor."* |
| `code.go` | 3. kapı — kod bloklarını `compile.go`'ya yönlendirir. |
| `compile.go` | `Verdict` üç değerli. *"ADDENDUM §B.2'nin dürüst yetenek sınırı, neden üç olduğunun sebebi: T0 'bu parse olmuyor'u ispatlayabilir ama 'bu, hiç verilmemiş bir classpath'e karşı semantik olarak doğru'yu asla ispatlayamaz, yani o durum `Skipped` — `Pass` değil, `Fail` de değil."* |
| `compile_bash.go` | `bash -n` — *"tek satır bile çalıştırmayan bir sözdizimi kontrolü. `bash -n`'in çözülmemiş bağımlılık kavramı yok (linking adımı yok), yani yaydığı her tanı bir sözdizimi hatası: her zaman `kindSyntax`, asla `kindUnresolved`."* |
| `compile_java.go` | *"`javac`'ı tek kullanımlık bir temp dizinde, JDK'nın ötesinde hiç classpath olmadan çalıştırır: Maven/Gradle çözümü yok, ağ yok, asla. Eksik bir kütüphane (ör. classpath'te Spring) 'package … does not exist' ya da 'cannot find symbol' verir — unresolved, snippet'te bir defekt değil."* |
| `compile_ts.go` | `tsUnresolvedCodes` = eksik import/`@types` paketinin ürettiği TS tanı kodları — *"bir sandbox-classpath sınırı, snippet'te bir defekt değil. Diğer her şey (özellikle TS1xxx parser-hata aralığı) gerçek bir problem."* |
| `freshness.go` | 4. kapı. Henüz tarihlenmemiş bir notta atlar — *"`schema.go`'nun `RetryOnce`'ı zaten o eksik-zorunlu-alan durumunu kapsıyor, yani `freshness.go` ikinci, yanıltıcı bir 'stale' verdict'i yığmak yerine atlıyor."* |
| `antislop.go` | 5. kapı. `structuralFail` `writing-rules.md`'nin Go'da zorlanan tek yapısal kuralını kontrol eder: *"howto/api notlarının bir gösterime ihtiyacı var, sadece bir iddiaya değil. Diğer beş tipin neden muaf olduğu için o dosyanın Structural requirements bölümüne bak."* Yasaklı-ifade taraması `vault.StripCode`'u kullanır. |
| `link.go` | 6. kapı — **yazmayı bloke etmez.** *"…bu `Fail` dürüstçe raporlanır ama `Report.Quarantine`'i set etmez (`gate.go`'nun `blocksWrite`'ı `DelegateToLibrarian`'ı dışlıyor) — not yine de `notes/`'a iner ve librarian'ın işi bir takip, bir tutma değil."* |
| `duplicate.go` | 7. kapı — bu da bloke etmez. Bir kullanıcının *"yine de aynı konuda iki not yayınlamak için belirtilmiş bir sebebi (DESIGN §12) olabilir."* |
| `quarantine.go` | `_inbox/` karantinası. `OpenQuestions` bir `Report`'un başarısız sonuçlarını `vault.WriteToInbox`'ın openQuestions madde işaretlerine çevirir — *"`Fail` başına bir tane, kapı sırasında; `Run`'ın `Outcomes`'ı kurduğu aynı sıra, böylece değişmemiş state üzerinde iki koşu byte-özdeş madde işaretleri üretir (B-020)."* |
| `gate_test.go`, `gate_adversarial_test.go`, `antislop_test.go`, `freshness_test.go`, `compile_test.go`, `quarantine_test.go`, `helpers_test.go` | Testler. `gate_adversarial_test.go` kapıları atlatmaya çalışan taslakları sürer. |

### 12.15 `pkg/sentinel/` (6) — yönetilen bloklar

| Dosya | Açıklama |
|---|---|
| `sentinel.go` | `Style` = bir yorum sözdizimi. *"`Close`, satır yorumları için boş (Go/Java `//`, Python `#`); Markdown'ın `<!--`/`-->`'ı işaret satırında iki uca da ihtiyaç duyuyor."* |
| `upsert.go` | *"`Upsert` dosyayı (ve üst dizinlerini) yoksa yaratır, mevcut bir id bloğunun gövdesini yerinde değiştirir, ya da dosya sonuna yeni bir blok ekler. Değişmemiş bir gövdeyle ikinci bir çağrı diske yeni hiçbir şey yazmaz."* Ayrıca `UpsertBefore`. |
| `remove.go` | `Remove` — `--remove-markers`'ın byte-bayt geri alımını mümkün kılan şey. |
| `lines.go` | `splitLines` eksik ya da boş bir dosyayı **sıfır satır** sayar, bir boş satır değil — *"taze bir dosyada `writeFile`'ın round trip'ini idempotent tutan şey budur."* |
| `write.go` | *"`writeFile` render edilmiş içerik zaten diskteki ile eşleştiğinde yazmayı atlar — özdeş bir yeniden yazım bile mtime'ı zıplatır, `cmd/forge`'un kendi yazarlarının yazmadan önce karşılaştırmasıyla aynı sebep. `Upsert`'ü diskte idempotent yapan şey o karşılaştırmadır, yalnızca içerikte değil."* |
| `sentinel_test.go` | Testler. |

**Kendi işaret çiftinin dışında hiçbir şeye dokunmaz.**

### 12.16 `pkg/dataset/` (10) — beş dataset

| Dosya | Açıklama |
|---|---|
| `d3.go` | *"ADDENDUM §D.1'in insan-düzeltmesi dataset'i: (model notu, senin düzenlediğin not) çiftleri, vault'un git geçmişinden hasat edilir. §D ona beşin en değerlisi diyor, ve Phase 6b yerine Phase 1'de inşa edilmesinin sebebi verinin yalnızca ileriye doğru birikmesi — sonradan kurulan bir hook, kurulmadan önce yapılmış düzenlemeleri kurtaramaz."* |
| `d2.go` | *"ADDENDUM §D.1'in advisor-distilasyon dataset'i: (taslak, kritik) çiftleri. D3 bir hook'un meşru olarak yeniden ateşleyebileceği bir git commit'inde dedupe ederken, D2'nin tetikleyicisi gerçek bir advisor çağrısı yapan tek bir `forge engine run` CLI çağrısı — korunacak bir yeniden-ateşleme yok, yani her yakalama kendi satırı olarak ekleniyor, `Key()` ya da idempotency kontrolü yok."* **B-024 kapandı** (2026-08-21): `D2Tag` artık `"d2"` — eskiden `"d2_advisor"` idi ve paketlenmiş config'in `"d2"` girdisiyle asla eşleşmediği için D2 shipped config altında sessizce inert kalıyordu. |
| `d4.go` | Karantina/taslak çiftleri. `D4Tag` paketlenmiş `dataset.capture` listesiyle birebir eşleşiyor (`"d4"`) — B-024'ün kapanmasıyla `D2Tag` de öyle. |
| `capture_gate_test.go` | **B-024'ün regression bekçisi ve bu bug'ın yeşil ship etmesinin sebebi:** hiçbir test config ile kodun uyuştuğunu iddia etmiyordu. `TestPackagedCaptureListGates` paketlenmiş katmanı tek başına yükleyip hem `Enabled` hem `D4Enabled`'ın `true` döndüğünü doğruluyor. Bilerek yalnızca D2 ve D4'ü kapsıyor — `d1`/`d3`/`d5` için kapı yok (**B-030**). `pkg/config`'e konamıyor: `dataset → vault → config` gerçek bir kenar, tersi cycle olurdu. |
| `d4_drafts.go` | *"…yazma başarısızlığı — tek dürüst join anahtarını tutar: kendisine geri verilen tam dosya."* |
| `git.go` | *"Tek bir repository'ye scope'lu, minimal salt-okunur git CLI kabuğu. D3 yakalaması bir git hook'unun içinde çalışıyor, yani git zaten path'te ve zaten bizi uyandıran process; altı plumbing okuması için go-git'i içeri çekmek getirisinden fazlaya mal olurdu. Kütüphane bağımlılığı `pkg/gitsig`'e ait (Phase 2b), ki onun blame ve churn'e ihtiyacı var."* |
| `jsonl.go` | *"`Append` dosyanın zaten tutmadığı çiftleri yazar ve kaç tane eklediğini döner. Dataset `.forge/` altında yaşıyor, ki türetilmiş ve vault'ta gitignore'lu, yani yerel kalıyor — ADDENDUM §D'nin dataset'leri hiçbir yere iletilmiyor."* |
| `d2_test.go`, `d3_test.go`, `d4_test.go`, `d3_forge_write_test.go` | Testler. Sonuncusu `Forge-Write: true` korumasını iki yönde de pinler (**B-007**). |

### 12.17 `pkg/telemetry/` (5)

| Dosya | Açıklama |
|---|---|
| `event.go` | `Event` DESIGN §14'ün şemasını alan alan yansıtır. *"json tag'leri dokümanın örnek satırıyla tam eşleşiyor, böylece `.forge/log.jsonl` o örneğe karşı gözle diff'lenebilir."* |
| `qhash.go` | *"`QHash` bir soruyu tek yönlü hash'ler, böylece log ham soru metnini asla taşımaz — DESIGN §14'ün invariant'ı. 12 hex karaktere kısaltılmış: tek bir vault'un ask hacminde tekrarları dedupe etmeye yetecek kadar, bir konunun ihtiyaç duyduğundan fazlasını saklamadan."* |
| `writer.go` | `Append` `<vaultDir>/.forge/log.jsonl`'e bir JSON satırı yazar. *"Saf I/O — çağıranlar bunu kendileri `cfg.Telemetry.Enabled` üzerinden kapılıyor, `engine_run.go`'nun `captureD2`'sinin `pkg/dataset`'in kendi config kontrolüne karşı aldığı aynı duruş."* |
| `qhash_test.go`, `writer_test.go` | Testler. |

### 12.18 `pkg/scrub/` (5)

| Dosya | Açıklama |
|---|---|
| `scrub.go` | `Report` bir `Scrub` koşusunu özetler: `NotesTotal`, `NotesWritten`, `Redactions`, `NoFrontmatter`. |
| `note.go` | *"`scrubOne` bir notu redakte eder. Hiç frontmatter bloğu olmayan bir not (birkaç gerçek notun hâlâ içinde olduğu migration-öncesi şekil) yalnızca-gövde redakte edilir; parse edilemeyen frontmatter'a sahip başka her şey şeklini tahmin etmek yerine **fails closed**."* |
| `redact.go` | `redact` her deseni sırayla uygular ve kaç değiştirme yaptığını raporlar. **İki gerçek false-positive sınıfı burada düzeltildi:** (1) kebab-case slug'lar ve tarihli dosya adları (`2026-04-13-local-ai-continue-rag-spring`) `sources:`/wikilink atıflarını bozuyordu → karakter sınıfından `-`/`_` düşürüldü; (2) gömülü kod örneklerindeki camelCase Java tanıtıcıları (`getPaymentOutboxMessageBySagaIdAndSagaStatus`) → eşleşmede **en az bir rakam** zorunlu kılındı; RE2'nin lookahead'i olmadığından filtre `redactLongTokens` içinde **eşleşme-sonrası** çalışır. `[A-Za-z0-9]`'dan rastgele 32+ karakterlik bir çekiliş neredeyse kesinlikle bir rakam içerir, yani gerçek hex/base64/JWT-biçimli sırlar hâlâ yakalanıyor. Gerçek vault'un 122 notunda redaksiyonlar **637 → 86 → 43**. |
| `write.go` | *"`writeAll` her scrub'lanmış notu diske işler. Yalnızca `scrubAll` tüm koşu için başarılı olduktan sonra çağrılır, böylece kısmi bir yazma asla olmaz."* |
| `scrub_test.go` | Fixture testi + iki regresyon testi: `TestScrubDoesNotRedactSlugsOrFilenames`, `TestScrubDoesNotRedactCamelCaseCodeIdentifiers`. |

> **Bilerek kovalanmayan bir kalıntı false positive:** içinde rakam barındıran bir tanıtıcı
> (ör. "E2E" üzerinden `TestE2ESessionContextRespectsTheBudget`) hâlâ heuristiği tetikliyor
> — ship edilen `examples/vault/`'ta bir örneği var.

---

## 13. `profiles/` (2 dosya)

| Dosya | Açıklama |
|---|---|
| `me.template.md` | Geliştirici profili şablonu. `forge init` bunu render edip `<vault>/profiles/me.md`'ye yazar. |
| `embed.go` | Şablonu binary'ye gömer. |

---

## 14. `references/` (6 dosya) — makine-okunur spec'ler

| Dosya | Açıklama |
|---|---|
| `schema.yaml` | **Not sözleşmesi.** `pkg/vault/schema.go` bunu okur; `forge validate`, `pkg/qualitygate/schema.go` ve `pkg/scrub`'ın fails-closed yeniden doğrulaması hep buna karşı çalışır. `check_render.go`'nun `values`'ı stack enum'unu buradan alır — *"rapor onsuz bir yokluğu isimlendiremez."* |
| `recall-spec.md` | **Phase 2'nin iki kararı burada argümanlanıyor ve sonraki fazlar bunu okumadan geri almamalı:** skor aktif kanallar üzerinden ağırlıklı **ortalama**, DESIGN §8'in literal ağırlıklı toplamı değil (§2.5); başlık ölçüsü **F₂, Dice değil** (§2.2). İkisi de ölçülmüş vault davranışından argümanlı. §3.1 kalibrasyon sweep'i ve tek açık defekti **B-008**. |
| `duplicate-spec.md` | Duplicate tespiti ve write-time-gate deseni. |
| `writing-rules.md` | Yazım kuralları. Structural requirements bölümü, `antislop.go`'nun neden yalnızca howto/api için gösterim zorunlu kıldığını ve diğer beş tipin neden muaf olduğunu açıklar. |
| `taxonomy.md` | Konu/stack taksonomisi. |
| `embed.go` | Hepsini binary'ye gömer. |

---

## 15. `scripts/` (4 dosya)

| Dosya | Açıklama |
|---|---|
| `install_vault_hook.sh` | `<vault-dir> [forge-binary]`. D3 capture hook'unu bir vault reposuna kurar. **Idempotent, ve ezmek yerine reddeder:** forge'a ait olmayan mevcut bir `post-commit` varsa durur. |
| `install_drift_hook.sh` | `<code-repo> <vault-dir> [forge-binary]`. Üç drift hook'unu bir kod reposuna kurar, bir vault'la eşleştirilmiş. Aynı idempotent/reddet duruşu. |
| `test-shim.sh` | `bin/forge` için fixture testi. *"Shim, vault'un sessiz post-commit hook'u ile doğrulanmamış bir binary arasında duran tek şey, yani 'benim makinemde çalışıyor' yeterli değil — fail-open bir shim, çalışan birinden ayırt edilemez olurdu."* |
| `migrate_vault.py` | **Hayatta kalan iki Python'dan biri** (diğeri offline dataset/fine-tuning araçları). Tek seferlik migration; binary'de ship etmez. Phase 1'de çalıştırıldı: 91 not taşındı, 345 wikilink yeniden yazıldı, 0 kırık, 60/91 şema-geçerli. |

---

## 16. `skills/` (4 dosya)

| Dosya | Açıklama |
|---|---|
| `forge/SKILL.md` | Ana dispatch. Soru → recall → pipeline → gate → yaz. Dört ürün agent'ına yönlendirir (bugün generic Agent tool + açık allowlist ile). |
| `forge-init/SKILL.md` | **Soruları soran taraf.** Beş soru (dil, framework'ler, infra, seniority, trigger) sorar ve `forge init`'e shell out eder. `depth` ve assume/never listeleri türetilir, sihirbaz beşte kalsın diye. |
| `forge-check/SKILL.md` | `/forge-check`. |
| `forge-stats/SKILL.md` | `/forge-stats`. |

Skill'ler iş mantığı taşımaz — soruları sorar ve binary'yi çağırır. Bilinçli: test
edilebilir mantık Go'da kalsın.

---

## 17. `templates/` (7 dosya)

Yedi not tipinin her biri için bir şablon: `concept.md`, `howto.md`, `api.md`,
`pattern.md`, `pitfall.md`, `incident.md`, `decision.md`.

Yedi tip, DESIGN §7'nin beş-dizinli ağacına karşı **B-005**'in kararıdır. Vault'ta yedi
`notes/<type>/` alt dizininin hepsi var, üçü boş `.gitkeep` kabuğu. **§7'ye uydurmak için
budama.**

Şablonlar `freshness_days`'in tip başına farklılığını da somutlaştırır: `api` 90 gün,
`howto` 180, `concept`/`pattern`/`pitfall` 365, `incident`/`decision` **0** (asla
bayatlamaz — geçmişte olan bir şeyi anlatıyorlar).

---

## 18. `testdata/` (16 dosya) — fixture vault

| Dosya | Açıklama |
|---|---|
| `README.md` | **F1–F12 defekt kataloğu.** Bunu okumadan fixture'a dokunma. |
| `vault/…` (13 not + `index.md`, `log.md`) | Gerçek vault'un **migration-öncesi** topolojisini yeniden üretir, artı on iki kasıtlı defekt: karışık frontmatter şekilleri, sarkan bir wikilink, sarkan bir `source:` yolu, bir orphan, bir near-duplicate çift, hiç frontmatter'ı olmayan notlar, gövde düzyazısında taşınan status. |

Dosyalar: `TIL/databases/keyset-cursor-pagination.md`,
`TIL/java-spring/saveandflush-timestamp-null.md`, `archive/old-rag-notes.md`,
`concepts/hibernate.md`, `concepts/soft-delete.md`, `concepts/soft-deletion.md`
(near-dupe çifti), `decisions/liquibase-over-column-alias.md`,
`entities/meter-readings-service.md`, `index.md`,
`issues/hibernate-column-mismatch.md`, `log.md`, `raw/daily/2026-04-13.md`,
`sources/daily/2026-04-13-local-ai-spring.md`,
`sources/daily/2026-04-14-spring-keycloak.md`, `syntheses/spring-persistence.md`.

> ### İki sert kural
>
> 1. **Defektler test yüzeyidir. DÜZELTME.**
> 2. **`.git`'i bilerek yok** — iç içe bir repo bu repo `git init`'lendiğinde başıboş bir
>    gitlink olurdu. Harness fixture'ı bir temp dizine kopyalar ve **kopyayı** `git init`
>    eder; migration'ın "kirli ağacı reddeder" ön koşulu ve drift'in `--since-commit`'i
>    böyle egzersiz edilir. **Asla yerinde `git init` etme.**
>
> Bir vault'u mutasyona uğratan **her şeyi önce burada prova et.** Phase 1'in migration'ı
> geri alınamazdı ve gerçek vault'un yedeği yoktu.

---

## 19. `examples/vault/` (93 dosya — grup olarak)

Phase 6'nın deliverable'ı. `forge scrub` gerçek vault'a karşı çalıştırılarak üretildi.

| Ne | Sayı |
|---|---|
| `notes/{pitfall,concept,decision,howto}/` altında notlar | 91 |
| `moc/codebase.md` | 1 |
| `moc/weekly/2026-W33.md` | 1 |

**Kapsam kararı:** yalnızca `notes/` + `moc/` — `raw/`, `sources/`, `reports/` ve
kök seviyesi başıboş dosyalar hariç. Bu, DESIGN §16.4'ün "15-20 gerçek not" satırını
**açık kullanıcı kararıyla geçersiz kılan** bir seçimdi. Kullanıcı scrub'lanmış çıktıyı
inceledi ve commit edilmeden önce onayladı — bu fazın kendi bağlayıcı review-gate
gereksinimi uyarınca.

`testdata/vault/` ile **karıştırma**: o 13 notluk kasıtlı-defektli bir fixture, bu ise
93 dosyalık temizlenmiş bir vitrin.

---

## 20. Dosya sayısı özeti

| Grup | Dosya |
|---|---|
| Kök | 12 |
| `.claude-plugin/` | 2 |
| `.claude/agents/` | 6 |
| `.github/workflows/` | 2 |
| `agents/` | 4 |
| `bin/` | 1 |
| `cmd/forge/` | 54 |
| `config/` | 10 |
| `docs/` | 10 |
| `evals/` | 7 |
| `hooks/` | 9 |
| `pkg/` (18 paket) | 160 |
| `profiles/` | 2 |
| `references/` | 6 |
| `scripts/` | 4 |
| `skills/` | 4 |
| `templates/` | 7 |
| `testdata/` | 16 |
| **Takip edilen toplam** | **316** |
| `examples/vault/` | +93 |
| **Genel toplam** | **409** |

`pkg/` dağılımı: `vault` 18 · `qualitygate` 20 · `report` 20 · `engine` 17 · `dataset` 10
· `recall` 9 · `drift` 9 · `codeindex` 8 · `config` 7 · `coderef` 6 · `sentinel` 6 ·
`store` 5 · `telemetry` 5 · `scrub` 5 · `graph` 4 · `gitsig` 4 · `linkcheck` 4 ·
`similarity` 3.

18 paketin hepsi `go test ./...`'te `ok` raporluyor (`config`, `profiles`, `references`
veri-only, test dosyası yok) — hem `CGO_ENABLED=0` hem `CGO_ENABLED=1` altında yeşil.
