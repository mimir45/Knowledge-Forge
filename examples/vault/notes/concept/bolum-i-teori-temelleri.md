---
title: 'BÖLÜM I: TEORİ TEMELLERİ'
slug: bolum-i-teori-temelleri
type: concept
depth: 3
confidence: low
created: 2026-08-09
updated: 2026-08-09
verified: 2026-08-09
freshness_days: 365
sources: []
related: []
supersedes: []
forge_version: 2.0.0
origin: import
---


---


	 
---

# BÖLÜM I: TEORİ TEMELLERİ

Bu bölüm, sınavı başarıyla geçmek için ihtiyacınız olan tüm teoriyi kapsar. Materyal, sınav alanlarına göre değil, teknolojilere ve kavramlara göre düzenlenmiştir — bu, her konuda daha derin bir anlayış oluşturmanıza yardımcı olur.

---

# Bölüm 1: Claude API — Model Etkileşiminin Temelleri

> Dokümantasyon: [Messages API](https://platform.claude.com/docs/en/api/messages) | [Prompt Engineering](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/overview)

## 1.1 API İstek Yapısı

Claude API, bir istek–yanıt (request–response) modelini izler. Claude Messages API'ye yapılan her istek şunları içerir:

```json
{
  "model": "claude-sonnet-4-6",
  "max_tokens": 1024,
  "system": "You are a helpful assistant.",
  "messages": [
    {"role": "user", "content": "Hi!"},
    {"role": "assistant", "content": "Hello!"},
    {"role": "user", "content": "How are you?"}
  ],
  "tools": [...],
  "tool_choice": {"type": "auto"}
}
```

**Temel alanlar:**
- `model` — model seçimi (`claude-opus-4-6`, `claude-sonnet-4-6`, `claude-haiku-4-5`)
- `max_tokens` — yanıttaki maksimum token sayısı
- `system` — system prompt (model davranışını tanımlar)
- `messages` — konuşma geçmişi (tutarlılığı korumak için **tüm geçmişi göndermeniz gerekir**)
- `tools` — kullanılabilir tool'ların tanımları
- `tool_choice` — tool seçim stratejisi

## 1.2 Mesaj Rolleri (Message Roles)

`messages` dizisi üç rol kullanır:
- `user` — kullanıcı mesajları
- `assistant` — model yanıtları (geçmiş gönderilirken dahil edilir)
- `tool` — tool çağrı sonuçları (rol açıkça ayarlanmaz; bu, bir `tool_result` content bloğu olarak görünür)

**Kritik derecede önemli:** her API isteğinde **tüm konuşma geçmişini** göndermelisiniz. Model, istekler arasında durum (state) tutmaz — her çağrı bağımsızdır.

## 1.3 Yanıttaki `stop_reason` Alanı

Claude API yanıtı, modelin neden üretmeyi durdurduğunu belirten `stop_reason` alanını içerir:

| Değer | Açıklama | Eylem |
|---|---|---|
| `"end_turn"` | Model yanıtını tamamladı | Sonucu kullanıcıya gösterin |
| `"tool_use"` | Model bir tool çağırmak istiyor | Tool'u çalıştırın ve sonucu geri döndürün |
| `"max_tokens"` | Token limitine ulaşıldı | Yanıt kesilmiştir; limiti artırmanız gerekebilir |
| `"stop_sequence"` | Bir stop sequence ile karşılaşıldı | Uygulama mantığınıza göre ele alın |

Agentic sistemler için en önemlileri `"tool_use"` ve `"end_turn"` değerleridir — agent döngüsünü (loop) bunlar kontrol eder.

## 1.4 System Prompt

System prompt, context'i ve davranış kurallarını tanımlayan özel bir talimattır. Şu özelliklere sahiptir:
- `messages` dizisinin parçası değildir; ayrı olarak `system` alanında iletilir
- Kullanıcı mesajlarına göre önceliklidir
- Bir kez yüklenir ve konuşma boyunca geçerlidir
- Rol, kısıtlamalar ve çıktı formatını tanımlamak için kullanılır

**Sınav için önemli:** system prompt ifadesi, istenmeyen tool ilişkileri oluşturabilir. Örneğin, "always verify the customer" gibi bir talimat, gereksiz olduğunda bile modelin `get_customer`'ı aşırı kullanmasına neden olabilir.

## 1.5 Context Window

Context window, modelin tek seferde işleyebileceği toplam metin miktarıdır (token cinsinden). Şunları içerir:
- System prompt
- Tüm mesaj geçmişi
- Tool tanımları
- Tool sonuçları

**Temel context window sorunları:**

1. **Lost-in-the-middle etkisi:** modeller, uzun bir girdinin başındaki ve sonundaki bilgiyi güvenilir şekilde işler ancak ortadaki detayları kaçırabilir. Çözüm: kilit bilgiyi başa veya sona yerleştirin.

2. **Tool sonuçlarının birikmesi:** her tool çağrısı, context'e çıktı ekler. Bir tool 40+ alan döndürüp yalnızca 5'i önemliyse, context'in çoğu boşa harcanır.

3. **Progressive summarization (kademeli özetleme):** geçmiş sıkıştırılırken sayısal değerler, yüzdeler ve tarihler sıklıkla kaybolur ve belirsizleşir ("yaklaşık", "kabaca", "birkaç").

---

# Bölüm 2: Tool'lar ve `tool_use`

> Dokümantasyon: [Tool Use](https://platform.claude.com/docs/en/build-with-claude/tool-use)

## 2.1 `tool_use` Nedir

`tool_use`, Claude'un harici fonksiyonları çağırmasına olanak tanıyan bir mekanizmadır. Model, kodu doğrudan çalıştırmaz — yapılandırılmış bir tool çağrı isteği üretir; sizin kodunuz bunu yürütür ve sonucu geri döndürür.

## 2.2 Tool Tanımı

Her tool, bir JSON şeması kullanılarak tanımlanır:

```json
{
  "name": "get_customer",
  "description": "Finds a customer by email or ID. Returns the customer profile, including name, email, order history, and account status. Use this tool BEFORE lookup_order to verify the customer's identity. Accepts an email (format: [REDACTED-EMAIL]) or a numeric customer_id.",
  "input_schema": {
    "type": "object",
    "properties": {
      "email": {"type": "string", "description": "Customer email"},
      "customer_id": {"type": "integer", "description": "Numeric customer ID"}
    },
    "required": []
  }
}
```

**Bir tool açıklamasının (description) kritik derecede önemli yönleri:**

1. **Açıklama, birincil seçim mekanizmasıdır.** Bir LLM, tool'ları açıklamalarına göre seçer. Minimal açıklamalar ("Retrieves customer information") tool'lar örtüştüğünde hatalara yol açar.

2. **Açıklamaya şunları ekleyin:**
   - Tool'un ne yaptığı ve ne döndürdüğü
   - Girdi formatları ve örnek değerler
   - Uç durumlar (edge cases) ve kısıtlamalar
   - Bu tool'un benzer alternatiflere karşı ne zaman kullanılacağı

3. **Tool'lar arasında** aynı veya örtüşen açıklamalardan **kaçının**. `analyze_content` ve `analyze_document` neredeyse aynı açıklamalara sahipse, model bunları karıştırır.

4. **Built-in tool'lar vs MCP tool'ları:** agent'lar, benzer işlevselliğe sahip MCP tool'ları yerine built-in tool'ları (Read, Grep) tercih edebilir. Bunu önlemek için MCP tool açıklamalarını güçlendirin — built-in tool'ların sağlayamayacağı somut avantajları, benzersiz verileri veya context'i vurgulayın.

## 2.3 `tool_choice` Parametresi

`tool_choice`, modelin tool'ları nasıl seçtiğini kontrol eder:

| Değer | Davranış | Ne zaman kullanılır |
|---|---|---|
| `{"type": "auto"}` | Model, bir tool çağırıp çağırmayacağına veya metinle yanıt vereceğine karar verir | Çoğu durum için varsayılan |
| `{"type": "any"}` | Model **mutlaka** bir tool çağırmalıdır | Garantili yapılandırılmış çıktıya ihtiyaç duyduğunuzda |
| `{"type": "tool", "name": "extract_metadata"}` | Model **mutlaka** belirli bir tool çağırmalıdır | Zorunlu bir ilk adıma / yürütme sırasına ihtiyaç duyduğunuzda |

**Önemli senaryolar:**
- `tool_choice: "any"` + birden çok extraction tool'u → model en iyisini seçer, ancak yine de yapılandırılmış çıktı alırsınız
- Zorunlu seçim → belirli bir ilk eylemi garantilemeniz gerektiğinde (örneğin, enrichment'tan önce `extract_metadata`)

## 2.4 Yapılandırılmış Çıktı için JSON Şemaları

JSON şemalarıyla `tool_use` kullanmak, Claude'dan yapılandırılmış çıktı elde etmenin **en güvenilir** yoludur. Şunları sağlar:
- Sözdizimsel olarak geçerli JSON garantisi (eksik süslü parantez yok, sondaki virgül yok)
- Gerekli yapının zorlanması (required alanlar mevcuttur)
- Anlamsal (semantic) doğruluğu **garanti etmez** (değerler yine de yanlış olabilir)

**Şema tasarımı — temel ilkeler:**

```json
{
  "type": "object",
  "properties": {
    "category": {
      "type": "string",
      "enum": ["bug", "feature", "docs", "unclear", "other"]
    },
    "category_detail": {
      "type": ["string", "null"],
      "description": "Details if category = 'other' or 'unclear'"
    },
    "severity": {
      "type": "string",
      "enum": ["critical", "high", "medium", "low"]
    },
    "confidence": {
      "type": "number",
      "minimum": 0,
      "maximum": 1
    },
    "optional_field": {
      "type": ["string", "null"],
      "description": "Null if the information was not found in the source"
    }
  },
  "required": ["category", "severity"]
}
```

**Şema tasarım kuralları:**
1. **Required vs optional:** bir alanı yalnızca bilgi her zaman mevcutsa required olarak işaretleyin. Required alanlar, veri eksik olduğunda modeli değer uydurmaya iter.
2. **Nullable alanlar:** mevcut olmayabilecek bilgiler için `"type": ["string", "null"]` kullanın. Model, halüsinasyon yerine `null` döndürebilir.
3. **`"other"` ile enum'lar:** önceden tanımlanmış kategorilerinizin dışındaki verileri kaybetmemek için `"other"` + bir detay string'i ekleyin.
4. **`"unclear"` enum'u:** modelin bir kategoriyi güvenle seçemediği durumlar için — dürüst bir `"unclear"`, yanlış bir kategoriden daha iyidir.

## 2.5 Sözdizimi (Syntax) vs Anlamsal (Semantic) Hatalar

| Hata tipi | Örnek | Çözüm |
|---|---|---|
| **Syntax** | Geçersiz JSON, yanlış alan tipi | JSON şemalı `tool_use` (ortadan kaldırır) |
| **Semantic** | Toplamlar tutmuyor, değer yanlış alanda, halüsinasyon | Doğrulama (validation) kontrolleri, geri bildirimle retry, self-correction |

---

# Bölüm 3: Claude Agent SDK — Agentic Sistemler Kurma

> Dokümantasyon: [Agent SDK](https://platform.claude.com/docs/en/agent-sdk/overview) | [Hooks](https://platform.claude.com/docs/en/agent-sdk/hooks) | [Subagents](https://platform.claude.com/docs/en/agent-sdk/subagents) | [Sessions](https://platform.claude.com/docs/en/agent-sdk/sessions)

## 3.1 Agentic Loop Nedir

Agentic loop, otonom görev yürütme için temel desendir. Model yalnızca yanıt vermez — bir dizi eylem gerçekleştirir:

```
1. Tool'larla birlikte Claude'a bir istek gönder
2. Bir yanıt al
3. stop_reason'ı kontrol et:
   - "tool_use" -> tool'u çalıştır, sonucu geçmişe ekle, 1. adıma geri dön
   - "end_turn" -> görev tamamlandı, sonucu kullanıcıya göster
4. Tamamlanana kadar tekrarla
```

**Bu, model-driven (model güdümlü) bir yaklaşımdır:** Claude, context'e ve önceki tool sonuçlarına dayanarak bir sonraki hangi tool'un çağrılacağına karar verir. Bu, eylem dizisinin sabit olduğu hard-coded karar ağaçlarından (decision tree) farklıdır.

**Anti-pattern'ler (kaçının):**
- Tamamlanmayı tespit etmek için assistant metnini parse etmek ("Task completed")
- Birincil durdurma koşulu olarak keyfi bir iterasyon limiti (örneğin `max_iterations=5`) kullanmak
- Tamamlanma sinyali olarak assistant'ın metinsel içerik üretip üretmediğini kontrol etmek

**Doğru yaklaşım:** tek güvenilir tamamlanma sinyali `stop_reason == "end_turn"`'dür.

## 3.2 `AgentDefinition` Yapılandırması

`AgentDefinition`, Claude Agent SDK'daki agent yapılandırma nesnesidir:

```python
agent = AgentDefinition(
    name="customer_support",
    description="Handles customer requests for returns and order issues",
    system_prompt="You are a customer support agent...",
    allowed_tools=["get_customer", "lookup_order", "process_refund", "escalate_to_human"],
    # Bir coordinator için:
    # allowed_tools=["Task", "get_customer", ...]
)
```

**Temel parametreler:**
- `name` / `description` — agent'ın kimliği ve açıklaması
- `system_prompt` — talimatları içeren system prompt
- `allowed_tools` — izin verilen tool'ların listesi (least privilege / en az ayrıcalık ilkesi)

## 3.3 Hub-and-Spoke: Coordinator ve Subagent'ler

Multi-agent mimarisi genellikle bir hub-and-spoke topolojisi olarak kurulur:

```
         Coordinator
        /     |      \
   Subagent1  Subagent2  Subagent3
    (search)   (analysis)   (synthesis)
```

**Coordinator şunlardan sorumludur:**
- Görevi alt görevlere (subtask) ayırmak
- Hangi subagent'lerin gerekli olduğuna karar vermek (dinamik seçim)
- İşi subagent'lere delege etmek
- Sonuçları toplamak (aggregate) ve doğrulamak
- Hataları ve retry'ları ele almak
- Sonuçları kullanıcıya iletmek

**Kritik ilke: subagent'lerin izole context'i vardır.**
- Subagent'ler, coordinator'ın konuşma geçmişini otomatik olarak **devralmaz**
- Gerekli tüm context, subagent prompt'unda **açıkça aktarılmalıdır**
- Subagent'ler çağrılar arasında bellek paylaşmaz
- Tüm iletişim coordinator üzerinden akar (gözlemlenebilirlik / observability ve hata kontrolü için)

## 3.4 Subagent Oluşturmak için `Task` Tool'u

Subagent'ler `Task` tool'u aracılığıyla oluşturulur (spawn):

```python
# Coordinator'ın allowedTools'u "Task" içermelidir
coordinator_agent = AgentDefinition(
    allowed_tools=["Task", "get_customer"]
)
```

**Açık context aktarımı zorunludur:**

```
# Kötü: subagent'in hiçbir context'i yok
Task: "Analyze the document"

# İyi: prompt'ta tam context
Task: "Analyze the following document.
Document: [full document text]
Prior search results: [web search results]
Output format requirements: [schema]"
```

**Paralel oluşturma (spawning):** bir coordinator, tek bir yanıtta birden çok `Task` çağırabilir — subagent'ler paralel çalışır:

```
# Bir coordinator yanıtı şunları içerir:
Task 1: "Search for articles about X"
Task 2: "Analyze document Y"
Task 3: "Search for articles about Z"
# Üçü de eş zamanlı (concurrent) çalışır
```

## 3.5 Agent SDK'da Hook'lar

Hook'lar, agent yaşam döngüsünün (lifecycle) belirli noktalarında müdahale (interception) ve dönüştürme (transformation) sağlar.

**PostToolUse**, bir tool sonucunu modele sağlanmadan önce yakalar:

```python
# Örnek: farklı MCP tool'larından gelen tarih formatlarını normalleştirme
@hook("PostToolUse")
def normalize_dates(tool_result):
    # Unix timestamp'i -> ISO 8601'e çevir
    # "Mar 5, 2025" -> "2025-03-05"e çevir
    return normalized_result
```

**Giden çağrıyı yakalayan hook**, politikayı ihlal eden eylemleri engeller:

```python
# Örnek: 500$ üzeri iadeleri engelle
@hook("PreToolUse")
def enforce_refund_limit(tool_call):
    if tool_call.name == "process_refund" and tool_call.args.amount > 500:
        return redirect_to_escalation(tool_call)
```

**Temel fark: hook'lar vs prompt talimatları**

| Özellik | Hook'lar | Prompt talimatları |
|---|---|---|
| Garanti | **Deterministik** (%100) | **Olasılıksal / Probabilistic** (>%90, %100 değil) |
| Ne zaman kullanılır | Kritik iş kuralları, finansal işlemler, uyumluluk (compliance) | Genel tercihler, öneriler, formatlama |
| Örnek | 500$ üzeri iadeleri engelle | "Escalate etmeden önce çözmeyi dene" |

**Kural:** başarısızlığın finansal, yasal veya güvenlik sonuçları olduğunda — prompt değil, hook kullanın.

# Bölüm 4: Model Context Protocol (MCP)

> Dokümantasyon: [MCP](https://modelcontextprotocol.io/) | [Tools](https://modelcontextprotocol.io/docs/concepts/tools) | [Resources](https://modelcontextprotocol.io/docs/concepts/resources) | [Servers](https://modelcontextprotocol.io/docs/concepts/servers)

## 4.1 MCP Nedir

Model Context Protocol (MCP), harici sistemleri Claude'a bağlamak için açık bir protokoldür. MCP, üç birincil resource tipi tanımlar:

1. **Tools** — agent'ın eylem gerçekleştirmek için çağırabileceği fonksiyonlar (CRUD işlemleri, API çağrıları, komut yürütme)
2. **Resources** — agent'ın context için okuyabileceği veriler (dokümantasyon, veritabanı şemaları, içerik katalogları)
3. **Prompts** — yaygın görevler için önceden tanımlanmış prompt şablonları

## 4.2 MCP Server'ları

Bir MCP server, MCP protokolünü uygulayan ve tool/resource sağlayan bir süreçtir (process). Bir MCP server'a bağlandığınızda:
- Tüm tool'lar otomatik olarak keşfedilir (discovered)
- Bağlı tüm server'lardan gelen tool'lar aynı anda kullanılabilir
- Tool açıklamaları, modelin bunları nasıl kullanacağını belirler

## 4.3 MCP Server'larını Yapılandırma

**Proje yapılandırması (`.mcp.json`)** — ekip kullanımı için:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "jira": {
      "command": "npx",
      "args": ["-y", "mcp-server-jira"],
      "env": {
        "JIRA_TOKEN": "${JIRA_TOKEN}"
      }
    }
  }
}
```

**Temel noktalar:**
- `.mcp.json`, proje kök dizininde saklanır ve version control'de yönetilir
- Sırlar (secrets) için environment variable'lar (`${GITHUB_TOKEN}`) kullanılır — token'ların kendisi commit edilmez
- Tüm proje katkıcılarına açıktır

**Kullanıcı yapılandırması (`~/.claude.json`)** — kişisel/deneysel server'lar için:
- Kullanıcının home dizininde saklanır
- Version control aracılığıyla paylaşılmaz
- Kişisel deneyler ve test için uygundur

**Server seçimi:**
- Standart entegrasyonlar için (Jira, GitHub, Slack), mevcut topluluk (community) MCP server'larını tercih edin
- Yalnızca benzersiz, ekibe özgü iş akışları için kendi server'larınızı oluşturun

## 4.4 MCP'de `isError` Bayrağı

Bir MCP tool bir hatayla karşılaştığında, yanıtta `isError: true` kullanır. Bu, agent'a çağrının başarısız olduğunu bildirir.

**Yapılandırılmış hata (iyi):**

```json
{
  "isError": true,
  "content": {
    "errorCategory": "transient",
    "isRetryable": true,
    "message": "The service is temporarily unavailable. Timeout while calling the orders API.",
    "attempted_query": "order_id=12345",
    "partial_results": null
  }
}
```

**Genel hata (anti-pattern):**

```json
{
  "isError": true,
  "content": "Operation failed"
}
```

Genel bir hata, agent'a karar verme için hiçbir bilgi vermez — retry mı yapmalı, sorguyu mu değiştirmeli, yoksa escalate mı etmeli?

## 4.5 MCP Resources

Resources, bir agent'ın eylem gerçekleştirmeden context elde etmek için talep edebileceği verilerdir:

- İçerik katalogları (örneğin, tüm proje görevlerinin listesi, hiyerarşik gezinme)
- Veritabanı şemaları (veri yapısını anlama)
- Dokümantasyon (API referansları, dahili kılavuzlar)
- Issue/görev özetleri

**Resource avantajı:** agent, hangi verinin mevcut olduğunu anlamak için keşif amaçlı (exploratory) tool çağrılarına ihtiyaç duymaz. Bir resource, anında bir "harita" sağlar.

---

# Bölüm 5: Claude Code — Yapılandırma ve İş Akışları

> Dokümantasyon: [Claude Code](https://code.claude.com/docs/en/overview) | [Memory / CLAUDE.md](https://code.claude.com/docs/en/memory) | [Skills](https://code.claude.com/docs/en/skills) | [MCP](https://code.claude.com/docs/en/mcp) | [Hooks](https://code.claude.com/docs/en/hooks) | [Sub-agents](https://code.claude.com/docs/en/sub-agents) | [GitHub Actions](https://code.claude.com/docs/en/github-actions) | [Headless](https://code.claude.com/docs/en/headless)

## 5.1 CLAUDE.md Hiyerarşisi

CLAUDE.md, Claude Code için talimat dosyalarıdır (instruction file). Üç seviyeli bir hiyerarşi vardır:

```
1. User-level: ~/.claude/CLAUDE.md
   - Yalnızca o kullanıcı için geçerlidir
   - VCS aracılığıyla paylaşılMAZ
   - Kişisel tercihler ve çalışma tarzı

2. Project-level: .claude/CLAUDE.md veya kök CLAUDE.md
   - Tüm proje katkıcıları için geçerlidir
   - VCS aracılığıyla yönetilir
   - Kodlama standartları, test standartları, mimari kararlar

3. Directory-level: alt dizinlerdeki CLAUDE.md
   - O dizindeki dosyalarla çalışırken geçerlidir
   - Kod tabanının o bölümüne özgü kurallar (conventions)
```

**Yaygın hata:** yeni bir ekip üyesi, proje talimatlarını almaz çünkü bu talimatlar `.claude/CLAUDE.md` (project-level) yerine `~/.claude/CLAUDE.md` (user-level) içine konmuştur.

## 5.2 `@path` Söz Dizimi (Dosya Import'ları)

CLAUDE.md, `@path` kullanarak harici dosyalara referans verebilir ve yapılandırmayı modüler hale getirebilir:

```markdown
# Project CLAUDE.md

Coding standards are described in @./standards/coding-style.md
Test requirements are in @./standards/testing-requirements.md
Project overview is in @README.md and dependencies are in @package.json
```

**`@path` kuralları:**
- Dosya yolundan hemen önce `@` kullanın (boşluk yok)
- Göreli (relative) ve mutlak (absolute) yollar desteklenir
- Göreli yollar, import'u içeren dosyaya göre çözümlenir
- Maksimum import iç içe geçme (nesting) derinliği 5'tir

Bu, tekrarı (duplication) önler ve her paketin yalnızca ilgili standartları içermesini sağlar.

## 5.3 `.claude/rules/` Dizini

`.claude/rules/`, monolitik bir CLAUDE.md'ye alternatiftir ve kuralları konuya göre düzenlemek için kullanılır:

```
.claude/rules/
  testing.md          -- test kuralları
  api-conventions.md  -- API kuralları
  deployment.md       -- deployment kuralları
  react-patterns.md   -- React desenleri
```

**Temel özellik: koşullu yükleme için `paths` ile YAML frontmatter:**

```yaml
---
paths: ["src/api/**/*"]
---

For API files, use async/await with explicit error handling.
Each endpoint must return a standard response wrapper.
```

```yaml
---
paths: ["**/*.test.tsx", "**/*.test.ts"]
---

Tests must use describe/it blocks.
Use data factories instead of hardcoding.
Do not mock the database—use a test database.
```

**Nasıl çalışır:**
- Bir kural, **yalnızca** Claude Code `paths` desenine uyan bir dosyayı düzenlediğinde yüklenir
- Bu, context ve token tasarrufu sağlar — ilgisiz kurallar yüklenmez
- Glob desenleri, konvansiyonları konumdan bağımsız olarak dosya tipine göre uygulamanızı sağlar (kod tabanına dağılmış testler için idealdir)

**`paths` ile `.claude/rules/` ne zaman vs directory-level CLAUDE.md ne zaman:**
- `paths` ile `.claude/rules/` — konvansiyonlar birçok dizine yayılmış dosyalara uygulandığında (testler, migration'lar)
- Directory-level CLAUDE.md — konvansiyonlar belirli bir dizine bağlı olduğunda ve başka yerde gerekmediğinde

## 5.4 Özel Slash Command'ler ve Skill'ler

> **Not:** mevcut Claude Code sürümünde, özel command'ler (`.claude/commands/`) skill'lerle (`.claude/skills/`) birleştirilmiştir. Her iki format da `/name` command'leri oluşturur. Sınav kılavuzu `.claude/commands/`'a atıfta bulunur — bu format hâlâ desteklenmektedir.

Slash command'ler, `/name` ile çağrılan yeniden kullanılabilir prompt şablonlarıdır:

**`.claude/commands/` formatı (legacy, destekleniyor):**

```
.claude/commands/
  review.md        -- /review -- standart kod incelemesi
  test-gen.md      -- /test-gen -- test üretimi
```

**`.claude/skills/` formatı (mevcut):**

```
.claude/skills/
  review/SKILL.md  -- /review -- frontmatter yapılandırmalı
  test-gen/SKILL.md
```

**Proje command'leri** (`.claude/commands/` veya `.claude/skills/`):
- VCS'de saklanır ve repo klonlandığında herkese açıktır
- Ekip genelinde tutarlı iş akışları sağlar

**Kullanıcı command'leri** (`~/.claude/commands/` veya `~/.claude/skills/`):
- VCS aracılığıyla paylaşılmayan kişisel command'ler
- Bireysel iş akışları için

## 5.5 Skill'ler — `.claude/skills/`

Skill'ler, SKILL.md frontmatter aracılığıyla yapılandırılan gelişmiş command'lerdir:

```yaml
---
context: fork
allowed-tools: ["Read", "Grep", "Glob"]
argument-hint: "Path to the directory to analyze"
---

Analyze the code structure in the specified directory.
Output a report on dependencies and architectural patterns.
```

**Frontmatter parametreleri:**

| Parametre | Açıklama |
|---|---|
| `context: fork` | Skill'i izole bir subagent'te çalıştırır. Ayrıntılı (verbose) çıktı ana oturumu (session) kirletmez |
| `allowed-tools` | Hangi tool'ların kullanılabileceğini kısıtlar (güvenlik — örneğin, izin verilmemişse skill dosyaları silemez) |
| `argument-hint` | Parametresiz çağrıldığında bir argüman isteyen ipucu |

**Skill ne zaman vs CLAUDE.md ne zaman:**
- **Skill** — belirli bir görev için isteğe bağlı (on-demand) çağrı (review, analiz, üretim)
- **CLAUDE.md** — her zaman yüklenen genel standartlar ve konvansiyonlar

**Kişisel skill'ler (`~/.claude/skills/`):**
- Ekip arkadaşlarınızı etkilememek için farklı adlar altında kişisel varyantlar oluşturun

## 5.6 Planning Mode vs Doğrudan Yürütme (Direct Execution)

**Planning mode:**
- Model yalnızca araştırır ve plan yapar; değişiklik yapmaz
- Kod tabanını keşfetmek için Read, Grep, Glob kullanır
- Kullanıcının onayladığı bir uygulama planı üretir
- Yan etkisi olmayan güvenli keşif

**Planning mode ne zaman kullanılır:**
- Büyük değişiklikler (düzinelerce dosya)
- Birden çok makul yaklaşım (microservices: sınırlar nasıl tanımlanır?)
- Mimari kararlar (hangi framework? hangi yapı?)
- Tanımadığınız bir kod tabanı (değiştirmeden önce anlamalısınız)
- 45+ dosyayı etkileyen kütüphane (library) migration'ları

**Doğrudan yürütme ne zaman kullanılır:**
- Net bir stack trace ile tek dosyalık düzeltmeler
- Tek bir doğrulama kontrolü ekleme
- İyi anlaşılmış, belirsizliği olmayan değişiklikler

**Birleşik yaklaşım:**
1. Araştırma ve tasarım için planning mode
2. Kullanıcı planı onaylar
3. Onaylanan planı uygulamak için doğrudan yürütme

**Explore subagent** — kod tabanını keşfetmek için özelleşmiş bir subagent:
- Ayrıntılı çıktıyı ana context'ten izole eder
- Yalnızca bir özet döndürür
- Çok aşamalı görevlerde context window tükenmesini önler

## 5.7 `/compact` Command'i

`/compact`, context'i sıkıştırmak için yerleşik (built-in) bir command'dir:
- Context window'u boşaltmak için önceki geçmişi özetler
- Context, ayrıntılı tool çıktısıyla dolduğunda uzun araştırma oturumlarında kullanılır
- Risk: özetleme sırasında kesin sayısal değerler, tarihler ve belirli detaylar kaybolabilir

## 5.8 `/memory` Command'i

`/memory`, oturumlar arası belleği yönetmek için yerleşik bir command'dir:
- Düzenleme için `CLAUDE.md` dosyasını açar ve notlar, tercihler ve context kaydetmenize olanak tanır
- Bilgi oturumlar arasında kalıcı olur ve başlangıçta otomatik olarak yüklenir
- Proje konvansiyonlarını, kullanıcı tercihlerini, sık kullanılan komutları ve mevcut çalışma context'ini saklamak için kullanışlıdır
- Aynı talimatları her oturumda yeniden açıklamaya alternatif

## 5.9 CI/CD için Claude Code CLI

**`-p` (veya `--print`) bayrağı:**

```bash
claude -p "Analyze this pull request for security issues"
```

- Non-interaktif mod: prompt'u işler, stdout'a yazdırır, çıkar
- Kullanıcı girdisini beklemez
- Claude'u CI/CD pipeline'larında çalıştırmanın tek doğru yolu

**CI için yapılandırılmış çıktı:**

```bash
claude -p "Review this PR" --output-format json --json-schema '{"type":"object",...}'
```

- `--output-format json` — JSON olarak çıktı
- `--json-schema` — çıktıyı bir şemaya göre doğrula
- Sonuç, inline PR yorumlarını otomatik olarak göndermek için parse edilebilir

**Oturum context izolasyonu:**
Kodu üreten aynı Claude oturumu, çoğu zaman onu incelemekte daha az etkilidir (model kendi muhakeme context'ini korur ve kendi kararlarını sorgulama olasılığı daha düşüktür). İnceleme için bağımsız bir instance kullanın.

**Yinelenen yorumları önleme:**
Yeni commit'ler sonrasında yeniden inceleme yaparken, önceki inceleme sonuçlarını context'e ekleyin ve Claude'a yalnızca yeni veya çözülmemiş sorunları bildirmesini söyleyin.

## 5.10 `fork_session` ve Oturum Yönetimi

**`--resume <session-name>`**, adlandırılmış bir oturumu sürdürür:

```bash
claude --resume investigation-auth-bug
```

- Kaydedilmiş context ile önceki bir konuşmayı sürdürür
- Birden çok oturuma yayılan uzun araştırmalar için kullanışlıdır
- Risk: dosyalar önceki oturumdan bu yana değiştiyse, tool sonuçları bayatlamış (stale) olabilir

**`fork_session`**, paylaşılan context'ten bağımsız bir dal (branch) oluşturur:

```
Codebase investigation
         |
    fork_session
    /           \
Approach A:      Approach B:
Redux            Context API
```

- Her iki fork da dallanma noktasına kadarki context'i devralır
- Daha sonra bağımsız olarak ayrışırlar
- Yaklaşımları veya test stratejilerini karşılaştırmak için kullanışlıdır

**Sürdürmek yerine ne zaman yeni bir oturum başlatmalı:**
- Tool sonuçları bayatlamış (dosyalar değişti)
- Çok zaman geçti ve context bozuldu
- Eski tool verisiyle sürdürmek yerine "İşte bulduklarımızın kısa bir özeti: ..." ile yeniden başlamak daha iyidir

---

# Bölüm 6: Prompt Engineering — İleri Teknikler

> Dokümantasyon: [Prompt Engineering](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/overview) | [Anthropic Cookbook](https://github.com/anthropics/anthropic-cookbook)

## 6.1 Few-shot Prompting

Few-shot prompting, beklenen davranışı göstermek için bir prompt'a 2–4 girdi/çıktı örneği eklenmesidir.

**Few-shot neden metinsel açıklamalardan daha etkilidir:**
- "be more precise" gibi belirsiz bir talimat birçok şekilde yorumlanabilir
- Bir örnek, beklenen formatı ve karar mantığını net biçimde gösterir
- Model, deseni yeni durumlara genelleştirir (yalnızca örnekleri tekrarlamaz)

**Few-shot örnek tipleri ve ne zaman kullanılacağı:**

1. **Belirsiz (ambiguous) senaryolar için örnekler:**

```
Request: "My order is broken"
Action: Call get_customer -> lookup_order -> check status.
Rationale: "broken" hasarlı bir ürün anlamına gelebilir; sipariş detaylarına ihtiyacınız var.

Request: "Get me a manager"
Action: Immediately call escalate_to_human.
Rationale: Müşteri açıkça bir insan istiyor. Otonom olarak çözmeye çalışmayın.
```

2. **Çıktı formatlama için örnekler:**

```
Finding example:
{
  "location": "src/auth/login.ts:42",
  "issue": "SQL injection in the username parameter",
  "severity": "critical",
  "suggested_fix": "Use a parameterized query"
}
```

3. **Kabul edilebilir vs sorunlu kodu ayırt eden örnekler:**

```
// Kabul edilebilir (flag'leme):
const items = data.filter(x => x.active);

// Sorun (flag'le):
const items = data.filter(x => x.active == true); // Use strict equality ===
```

4. **Farklı belge formatlarından çıkarım için örnekler:**

```
Document with inline citations:
"As shown in the study (Smith, 2023), the rate is 42%."
-> {"value": "42%", "source": "Smith, 2023", "type": "inline_citation"}

Document with bibliography references:
"The rate is 42%. [1]"
-> {"value": "42%", "source": "reference_1", "type": "bibliography"}
```

5. **Gayriresmî (informal) ölçümler için örnekler:**

```
Text: "about two handfuls of rice"
-> {"amount": "~100g", "original_text": "two handfuls", "precision": "approximate"}

Text: "a pinch of salt"
-> {"amount": "~1g", "original_text": "a pinch", "precision": "approximate"}
```

Few-shot, salt kural tabanlı talimatlar için fazla çeşitli olan gayriresmî ve standart dışı ölçü birimlerini çıkarmada özellikle etkilidir.

**Prompt'larda format normalleştirme kuralları:**
Yapılandırılmış çıktı için katı (strict) JSON şemaları kullanırken, prompt'a normalleştirme kuralları ekleyin:

```
Normalization:
- Dates: always ISO 8601 (YYYY-MM-DD); "yesterday" -> compute an absolute date
- Currency: numeric amount + currency code; "five bucks" -> {"amount": 5, "currency": "USD"}
- Percentages: decimal fraction; "half" -> 0.5
```

Bu, JSON'un sözdizimsel olarak geçerli olduğu ancak değerlerin tutarsız olduğu anlamsal hataları önler.

## 6.2 Açık Kriterler (Explicit Criteria) vs Belirsiz Talimatlar

**Kötü (belirsiz):**

```
Check code comments for accuracy.
Be conservative—report only high-confidence findings.
```

**İyi (açık kriterler):**

```
Flag a comment as problematic ONLY if:
1. The comment describes behavior that CONTRADICTS the actual code behavior
2. The comment references a non-existent function or variable
3. A TODO/FIXME comment refers to a bug that has already been fixed in code

Do NOT flag:
- Comments that are merely stylistically outdated
- Comments with minor wording inaccuracies
- Missing comments (that is a separate category)
```

**Severity kriterlerini örneklerle tanımlayın:**

```
CRITICAL: Runtime failure for users
  Example: NullPointerException while processing a payment

HIGH: Security vulnerability
  Example: SQL injection, XSS, missing authorization checks

MEDIUM: Logic bug without immediate impact
  Example: Wrong sorting, off-by-one error

LOW: Code quality
  Example: Duplication, suboptimal algorithm for small data
```

## 6.3 Prompt Chaining

Prompt chaining, karmaşık bir görevi odaklanmış adımlardan oluşan bir diziye böler:

```
Step 1: Analyze auth.ts (yalnızca yerel sorunlar)
       -> Output: auth.ts içindeki sorunların listesi

Step 2: Analyze database.ts (yalnızca yerel sorunlar)
       -> Output: database.ts içindeki sorunların listesi

Step 3: Integration pass (dosyalar arası bağımlılıklar)
       -> Output: modül sınırlarındaki sorunlar
```

**Bunun önemi:**
- **Attention dilution'ı (dikkat dağılması)** önler — model aynı anda çok fazla dosya aldığında, bazı dosyalardaki bug'ları kaçırabilir ve yüzeysel yorumlar yapabilir
- Dosya başına tutarlı analiz kalitesi sağlar
- Dosyalar arası etkileşimlerin ayrı analizine olanak tanır

**Prompt chaining ne zaman vs dynamic decomposition ne zaman:**
- **Prompt chaining** — öngörülebilir, tekrarlanabilir görevler (kod incelemesi, dosya migration'ları)
- **Dynamic decomposition** — alt görevlerin yalnızca yürütme sırasında netleştiği açık uçlu araştırmalar

## 6.4 "Interview" (Mülakat) Deseni

Bir çözümü uygulamadan önce, Claude açıklayıcı sorular sorar:

```
Claude: "Before implementing caching for the API, a few questions:
1. Which cache invalidation strategy do you prefer—TTL or event-based?
2. Is stale data acceptable when the cache is unavailable?
3. Should caching be per-user or global?
4. What is the expected data volume to cache?"
```

**Bunun yararlı olduğu durumlar:**
- Tanımadığınız bir alan (fintech, sağlık, hukuk sistemleri)
- Açık olmayan sonuçları olan görevler (cache stratejileri, failure mode'ları)
- En iyi seçimin context'e bağlı olduğu birden çok uygulanabilir yaklaşım

## 6.5 Doğrulama ve Geri Bildirimle Retry (Validation and Retry-with-Feedback)

Çıkarılan veri doğrulamadan geçemediğinde:

```
Step 1: Belgeden veri çıkar
Step 2: Doğrula (Pydantic, JSON Schema, iş kuralları)
Step 3: Bir hata varsa—context ile retry:
  - Orijinal belge
  - Önceki (hatalı) çıkarım
  - Belirli hata: "Field 'total' = 150, but sum(line_items) = 145. Re-check values."
```

**Retry'nin etkili olacağı durumlar:**
- Format hataları (yanlış formatta tarih)
- Yapısal hatalar (yanlış konuma yerleştirilmiş bir alan)
- Aritmetik tutarsızlıklar (model yeniden kontrol edebilir)

**Retry'nin YARDIMCI OLMAYACAĞI durumlar:**
- Bilgi kaynak belgede yoktur
- Gerekli context harici kaynaktadır (veri sağlanmamış başka bir belgededir)

**Doğrulama tool'u olarak Pydantic:**
Pydantic, şema tabanlı veri doğrulaması için bir Python kütüphanesidir. Sınav için temel noktalar:
- **Yapısal doğrulama:** Claude'dan JSON alındıktan sonra kodda kontrol edilen tipler, requiredlık, enum kısıtlamaları
- **Anlamsal doğrulama:** custom validator'lar iş mantığını uygular (kalemlerin toplamı total'a eşit; start_date < end_date)
- **Validate–retry döngüleri:** Pydantic doğrulaması başarısız olduğunda, bir hata mesajı oluşturun ve hata context'i ile Claude'a yeniden prompt verin
- **JSON Schema üretimi:** Pydantic modelleri, `tool_use` için JSON Schema üretebilir ve tek bir doğruluk kaynağı (single source of truth) sağlar

## 6.6 Self-correction (Öz Düzeltme)

İç çelişkileri tespit etmek için bir desen:

```json
{
  "stated_total": "$150.00",
  "calculated_total": "$145.00",
  "conflict_detected": true,
  "line_items": [
    {"name": "Widget A", "price": 75.00},
    {"name": "Widget B", "price": 70.00}
  ]
}
```

Model hem belirtilen (stated) değeri hem de hesaplanan (computed) bir değeri çıkarır — bunlar farklıysa, `conflict_detected` tutarsızlığı ele almanıza olanak tanır.

---

# Bölüm 7: Message Batches API

> Dokümantasyon: [Message Batches](https://platform.claude.com/docs/en/build-with-claude/message-batches)

## 7.1 Genel Bakış

Message Batches API, asenkron işleme için istek grupları (batch) göndermenize olanak tanır:

| Özellik | Değer |
|---|---|
| Tasarruf | Senkron çağrılara kıyasla **%50** |
| İşleme penceresi | **24 saate** kadar (latency SLA garantisi yok) |
| Multi-turn tool calling | **Desteklenmez** (bir istek = bir yanıt) |
| Korelasyon | İstek ve yanıtı bağlamak için `custom_id` alanı |

## 7.2 Batch API ne zaman vs Senkron API ne zaman

| Görev | API | Neden |
|---|---|---|
| Pre-merge PR kontrolü | **Senkron** | Geliştirici bekliyor; 24 saat kabul edilemez |
| Gece çalışan tech-debt raporu | **Batch** | Sonuç sabaha lazım; %50 tasarruf |
| Haftalık güvenlik denetimi | **Batch** | Acil değil; %50 tasarruf |
| İnteraktif kod incelemesi | **Senkron** | Anında yanıt gerekli |
| 10.000 belge işleme | **Batch** | Toplu işleme; tasarruf önemli |

## 7.3 `custom_id` Kullanımı

```json
{
  "custom_id": "doc-invoice-2024-001",
  "params": {
    "model": "claude-sonnet-4-6",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Extract data from: ..."}]
  }
}
```

`custom_id` şunları yapmanıza olanak tanır:
- Sonucu orijinal belgeye bağlama
- Başarısızlık durumunda yalnızca başarısız belgeleri yeniden gönderme
- Başarılı belgeleri yeniden işlemekten kaçınma

## 7.4 Batch'lerde Başarısızlıkları Ele Alma

1. 100 belgelik bir batch gönder
2. 95'i başarılı; 5'i başarısız (context limiti aşıldı)
3. Başarısızlıkları `custom_id` ile tespit et
4. Stratejiyi değiştir (örneğin, uzun belgeleri chunk'lara böl)
5. Yalnızca başarısız olan 5 belgeyi yeniden gönder

## 7.5 SLA Planlaması

30 saatte bir sonuca ihtiyacınız varsa ve Batch API 24 saate kadar sürebiliyorsa:
- Gönderim penceresi: 30 - 24 = **6 saat**
- Batch'ler, son teslim tarihinden en geç 24 saat önce gönderilmelidir
- Sık gönderimler için, 4 saatlik pencerelere bölün

---

# Bölüm 8: Görev Ayrıştırma (Task Decomposition) Stratejileri

## 8.1 Sabit Pipeline'lar (Prompt Chaining)

Her adım önceden tanımlanır:

```
Document -> Metadata extraction -> Data extraction -> Validation -> Enrichment -> Final output
```

**Ne zaman kullanılır:**
- Görev yapısı öngörülebilir (incelemeler her zaman aynı şablonu izler)
- Tüm adımlar baştan bilinir
- Kararlılığa (stability) ve tekrarlanabilirliğe ihtiyacınız var

## 8.2 Dinamik Uyarlanabilir Ayrıştırma (Dynamic Adaptive Decomposition)

Alt görevler, ara sonuçlara dayanarak üretilir:

```
1. "Add tests for a legacy codebase"
2. -> İlk: yapıyı haritala (Glob, Grep)
3. -> Bulundu: testi olmayan 3 modül, kısmi kapsamlı 2 modül
4. -> Önceliklendir: payments modülüyle başla (yüksek risk)
5. -> Çalışırken: harici bir API'ye bağımlılık keşfedildi
6. -> Uyarla: test yazmadan önce harici API için bir mock ekle
```

**Ne zaman kullanılır:**
- Açık uçlu araştırma görevleri
- Tam kapsam baştan bilinmediğinde
- Her adım, öncekinin sonuçlarına bağlı olduğunda

## 8.3 Çok Geçişli (Multi-pass) Kod İncelemesi

10+ dosyalı pull request'ler için:

```
Pass 1 (dosya başına): Analyze auth.ts -> yerel sorunları listele
Pass 1 (dosya başına): Analyze database.ts -> yerel sorunları listele
Pass 1 (dosya başına): Analyze routes.ts -> yerel sorunları listele
...
Pass 2 (entegrasyon): Dosyalar arası ilişkileri analiz et
  -> Dosyalar arası sorunlar: tutarsız tipler, döngüsel bağımlılıklar
```

**14 dosya üzerinden tek geçiş neden kötüdür:**
- Attention dilution: bazı dosyalar için derin, bazıları için yüzeysel analiz
- Tutarsız yorumlar: bir desen bir dosyada flag'lenir ama başka bir dosyada onaylanır
- Kaçırılan bug'lar: bilişsel aşırı yük nedeniyle bariz hatalar atlanır

---

# Bölüm 9: Escalation ve Human-in-the-Loop

## 9.1 Ne Zaman Bir İnsana Escalate Edilir

**Escalation tetikleyicileri (net kurallar):**

| Durum | Eylem |
|---|---|
| Müşteri açıkça "get me a manager" diyor | Hemen escalate et; çözmeye çalışma |
| Politika talebi kapsamıyor | Escalate et (örneğin, politika sessizken rakip fiyat eşleştirme) |
| Agent ilerleme kaydedemiyor | Makul sayıda denemeden sonra escalate et |
| Bir eşiğin üzerinde finansal işlem | Escalate et (tercihen prompt değil, bir hook ile zorunlu kılınmalı) |
| Müşteri ararken birden çok eşleşme | Ek tanımlayıcı (identifier) iste; tahmin etme |

**Güvenilir bir tetikleyici OLMAYAN şeyler:**

| Güvenilmez yöntem | Neden başarısız olur |
|---|---|
| Duygu analizi (sentiment analysis) | Müşteri ruh hali, vaka karmaşıklığıyla ilişkili değildir |
| Modelin kendi öz değerlendirme güveni (1–10) | Model emin bir şekilde yanlış olabilir; kalibrasyon zayıftır |
| Otomatik bir sınıflandırıcı (classifier) | Aşırı mühendislik; sahip olmayabileceğiniz eğitim verisi gerektirebilir |

## 9.2 Escalation Desenleri

**Anında escalation:**

```
Customer: "I want to speak to a manager"
Agent: [hemen escalate_to_human çağırır]
DEĞİL: "I can help with your issue, let me..."
```

**Çözüm denemesinden sonra escalation:**

```
Customer: "My refrigerator broke two days after purchase"
Agent: [siparişi kontrol eder, garanti değişimi önerir]
M�şteri memnun değilse -> escalate et
```

**Nüanslı escalation (kabul et → çöz → tekrarlanırsa escalate et):**

```
Customer: "This is outrageous, I'm very unhappy with the quality!"
Agent: [hayal kırıklığını kabul eder] "I understand your frustration."
       [çözüm sunar] "I can offer a replacement or a refund."
Customer: "No, I want to talk to someone!"
Agent: [müşteri tekrar ısrar ediyor -> anında escalation]
```

Temel ilke: önce duyguyu kabul et, sonra somut bir çözüm öner ve yalnızca müşteri bir insan isteğini tekrarlarsa escalate et. İlk memnuniyetsizlik ifadesinde escalate etme (bu, bir yönetici istemekle aynı şey değildir).

**Politika boşluğu (policy gap) için escalation:**

```
Customer: "Competitor X has this item 30% cheaper—give me a discount"
Policy: yalnızca kendi sitenizde fiyat ayarlamalarını kapsıyor
Agent: [escalate eder — politika rakip fiyat eşleştirmeyi kapsamıyor]
```

## 9.3 Yapılandırılmış Handoff (Devir) Protokolleri

Escalation'da agent, bir insana yapılandırılmış bir özet iletmelidir:

```json
{
  "customer_id": "CUST-12345",
  "customer_name": "Ivan Petrov",
  "issue_summary": "Refund request for a damaged item",
  "order_id": "ORD-67890",
  "root_cause": "Item arrived damaged; photos attached",
  "actions_taken": [
    "Verified customer via get_customer",
    "Confirmed order via lookup_order",
    "Offered a standard replacement — customer insists on a refund"
  ],
  "refund_amount": "$89.99",
  "recommended_action": "Approve a full refund",
  "escalation_reason": "Customer requested to speak with a manager"
}
```

İnsan operatör, tam konuşma transkriptine erişemez — yalnızca bu özeti görür. Bu nedenle özet eksiksiz ve kendi kendine yeterli (self-contained) olmalıdır.

## 9.4 Güven Kalibrasyonu (Confidence Calibration) ve İnsan Gözetimi

Veri çıkarma sistemleri için:

1. **Alan bazında (field-level) güven skorları:** model, çıkarılan her alan için bir güven skoru üretir
2. **Kalibrasyon:** eşikleri ayarlamak için etiketli (labeled) doğrulama setleri kullanın
3. **Yönlendirme (routing):**
   - Yüksek güven + kararlı doğruluk -> otomatik işleme
   - Düşük güven veya belirsiz kaynaklar -> insan incelemesi

**Tabakalı rastgele örnekleme (stratified random sampling):**
- Yüksek güvenli çıkarımlar için bile, düzenli olarak bir örnek denetleyin
- Toplam %97 doğruluk, belirli bir belge tipi için %40 hatayı gizleyebilir
- Doğruluğu yalnızca genel olarak değil, belge tipine ve alana göre analiz edin

---

# Bölüm 10: Multi-agent Sistemlerde Hata Yönetimi (Error Handling)

## 10.1 Hata Kategorileri

| Kategori | Örnekler | Retry edilebilir mi | Agent eylemi |
|---|---|---|---|
| **Transient** | Timeout, 503, ağ hatası | Evet | Exponential backoff ile retry |
| **Validation** | Geçersiz girdi formatı, eksik required alan | Hayır (girdiyi düzelt) | İsteği değiştir ve retry et |
| **Business** | Politika ihlali, eşik aşımı | Hayır | Kullanıcıya açıkla; bir alternatif öner |
| **Permission** | Erişim reddedildi | Hayır | Escalate et |

## 10.2 Hata Yönetimi Anti-pattern'leri

| Anti-pattern | Sorun | Doğru yaklaşım |
|---|---|---|
| Genel "search unavailable" durumu | Coordinator nasıl kurtarılacağına karar veremez | Hata tipini, sorguyu, kısmi sonuçları, alternatifleri döndür |
| Sessiz bastırma (boş sonuç = başarı) | Coordinator eşleşme olmadığını sanır ama aslında bir hata oldu | "Sonuç yok" ile "arama hatası"nı ayırt et |
| Tek başarısızlıkta tüm iş akışını iptal etme | Tüm kısmi sonuçları kaybedersiniz | Kısmi sonuçlarla devam et; boşlukları annotate et |
| Bir subagent içinde sonsuz retry | Latency ve kaynak israfı | Yerel kurtarma (1–2 retry), sonra coordinator'a iletme |

## 10.3 Yapılandırılmış Bir Subagent Hatası

```json
{
  "status": "partial_failure",
  "failure_type": "timeout",
  "attempted_query": "AI impact on music industry 2024",
  "partial_results": [
    {"title": "AI Music Generation Report", "url": "...", "relevance": 0.8}
  ],
  "alternative_approaches": [
    "Try a narrower query: 'AI music composition tools'",
    "Use an alternative data source"
  ],
  "coverage_impact": "Not covered: AI impact on music production"
}
```

Bu, coordinator'a karar vermek için gereken bilgiyi sağlar:
- Değiştirilmiş bir sorguyla retry et?
- Kısmi sonuçları kullan?
- Farklı bir subagent'e delege et?
- Bu bölüm olmadan devam et ve boşluğu annotate et?

## 10.4 Nihai Sentezdeki Kapsam (Coverage) Notları

```markdown
## Report: AI Impact on Creative Industries

### Visual Art (FULL COVERAGE)
[research results]

### Music (PARTIAL COVERAGE — search agent timeout)
[partial results]
⚠️ Note: coverage for this section is limited due to a timeout in the search agent.

### Literature (FULL COVERAGE)
[research results]
```

---

# Bölüm 11: Production Sistemlerinde Context Yönetimi

## 11.1 Olguları (Facts) Ayrı Bir Bloğa Çıkarma

Özetleme sırasında bozulan konuşma geçmişine güvenmek yerine, kilit olguları yapılandırılmış bir bloğa çıkarın:

```
=== CASE FACTS (yeni bir olgu ortaya çıktıkça güncellenir) ===
Customer ID: CUST-12345
Order ID: ORD-67890
Order Date: 2025-01-15
Order Amount: $89.99
Issue: Damaged item on delivery
Customer Request: Full refund
Status: Pending manager approval
===
```

Geçmiş nasıl özetlenirse özetlensin, bu bloğu her prompt'a dahil edin.

## 11.2 Tool Sonuçlarını Kırpma (Trimming)

`lookup_order` 40+ alan döndürüyor ama mevcut görev için yalnızca 5'ine ihtiyacınız varsa:

```python
# PostToolUse hook: yalnızca ilgili alanları tut
@hook("PostToolUse", tool="lookup_order")
def trim_order_fields(result):
    return {
        "order_id": result["order_id"],
        "status": result["status"],
        "total": result["total"],
        "items": result["items"],
        "return_eligible": result["return_eligible"]
    }
```

Bu, context tasarrufu sağlar ve gürültüyü (noise) azaltır.

## 11.3 Konuma Duyarlı (Position-aware) Girdi

Kritik bilgiyi lost-in-the-middle etkisini göz önünde bulundurarak yerleştirin:

```
[KEY FINDINGS — en üstte]
Found 3 critical vulnerabilities...

[DETAILED RESULTS — ortada]
=== File auth.ts ===
...
=== File database.ts ===
...

[ACTION ITEMS — en sonda]
Priority: fix auth.ts vulnerabilities before merge.
```

## 11.4 Scratchpad Dosyaları

Uzun araştırmalarda, agent kilit bulguları bir scratchpad dosyasına yazabilir:

```
# investigation-scratchpad.md
## Key findings
- PaymentProcessor in src/payments/processor.ts inherits from BaseProcessor
- refund() is called from 3 places: OrderController, AdminPanel, CronJob
- External PaymentGateway API has a rate limit of 100 req/min
- Migration #47 added refund_reason (NOT NULL) — 2024-12-01
```

Context bozulduğunda (veya yeni bir oturumda), agent keşfi yeniden çalıştırmak yerine scratchpad'e başvurabilir.

## 11.5 Context'i Korumak için Subagent'lere Delege Etme

```
Main agent: "Investigate dependencies of the payments module"
  -> Subagent (Explore): 15 dosya okur, import'ları izler
  -> Returns: "Payments depends on AuthService, OrderModel, and the external PaymentGateway API"

Main agent: 15 dosya yerine context'te tek bir satır tutar
```

**Ayrı context katmanı (separate context layer):**
Multi-agent sistemlerde, her subagent sınırlı bir context bütçesi içinde çalışır — yalnızca görevi için gereken bilgiyi alır. Coordinator, ayrı bir context katmanı görevi görür: subagent çıktılarını toplar, global state'i saklar ve context tahsis eder. Bu, bir agent'ın diğerleri için ilgisiz bilgilerle pencereyi tükettiği "context sızıntısını" (context leakage) önler.

**Subagent'ler için kısıtlanmış context bütçeleri:**
- Minimal context gönderin: belirli bir görev + gerekli veri
- Subagent'e ham veri yığınları değil, yapılandırılmış sonuçlar döndürmesini söyleyin
- Subagent'in tool setini sınırlamak için `allowedTools` kullanın — daha az tool, daha az dikkat dağınıklığı ve daha düşük context maliyeti demektir

## 11.6 Yapılandırılmış State Kalıcılığı (Crash Recovery için)

Her agent, durumunu bilinen bir konuma export eder:

```json
// agent-state/web-search-agent.json
{
  "status": "completed",
  "queries_executed": ["AI music 2024", "AI music composition"],
  "results_count": 12,
  "key_findings": [...],
  "coverage": ["music composition", "music production"],
  "gaps": ["music distribution", "music licensing"]
}
```

Coordinator, sürdürme (resume) sırasında bir manifest yükler:

```json
// agent-state/manifest.json
{
  "web-search": "completed",
  "doc-analysis": "in_progress",
  "synthesis": "not_started"
}
```

---

# Bölüm 12: Provenance'ı (Kaynak/Köken) Koruma

## 12.1 Atıf Kaybı (Attribution Loss) Sorunu

Birden çok kaynaktan gelen sonuçlar özetlenirken, "claim → source" (iddia → kaynak) bağlantısı kaybolabilir:

```
Kötü: "The AI music market is estimated at $3.2B." (Kaynak yok, yıl yok.)

İyi:
{
  "claim": "The AI music market is estimated at $3.2B.",
  "source_url": "https://example.com/report",
  "source_name": "Global AI Music Report 2024",
  "publication_date": "2024-06-15",
  "confidence": 0.9
}
```

## 12.2 Çelişkili Verileri Ele Alma

İki kaynak farklı değerler verdiğinde:

```json
{
  "claim": "Share of AI-generated music on streaming platforms",
  "values": [
    {
      "value": "12%",
      "source": "Spotify Annual Report 2024",
      "date": "2024-03",
      "methodology": "Automated classification"
    },
    {
      "value": "8%",
      "source": "Music Industry Association Survey",
      "date": "2024-07",
      "methodology": "Survey of 500 labels"
    }
  ],
  "conflict_detected": true,
  "possible_explanation": "Difference in methodology and time period"
}
```

Keyfi olarak bir değer seçmeyin. Her ikisini de atıfla koruyun ve kararı coordinator'a bırakın.

## 12.3 Doğru Yorumlama için Tarihleri Dahil Edin

Tarihler olmadan, zamansal farklılıklar yanlışlıkla çelişki olarak yorumlanabilir:

```
Kötü: "Source A says 10%, source B says 15%. Contradiction."
İyi: "Source A (2023) says 10%, source B (2024) says 15%. Likely +5% growth over a year."
```

## 12.4 İçerik Tipine Göre Render Edin

Her şeyi tek bir formata zorlamayın:
- Finansal veri -> tablolar
- Haber ve analiz -> düz metin (prose)
- Teknik bulgular -> yapılandırılmış listeler
- Zaman serisi -> kronolojik sıralama

---

# Bölüm 13: Claude Code Built-in Tool'ları

## 13.1 Tool Seçim Referansı

| Görev | Tool | Örnek |
|---|---|---|
| Ada/desene göre dosya bul | **Glob** | `**/*.test.tsx`, `src/components/**/*.ts` |
| Dosyalar içinde ara | **Grep** | Fonksiyon adı, hata mesajı, import |
| Bir dosyayı tam olarak oku | **Read** | Analiz için bir dosya yükle |
| Yeni bir dosya yaz | **Write** | Sıfırdan bir dosya oluştur |
| Mevcut bir dosyayı kesin olarak düzenle | **Edit** | Benzersiz metin eşleşmesi ile belirli bir snippet'i değiştir |
| Bir shell komutu çalıştır | **Bash** | git, npm, test çalıştır, build |

## 13.2 Kademeli (Incremental) Araştırma Stratejisi

Tüm dosyaları aynı anda okumayın. Anlayışı kademeli olarak oluşturun:

```
1. Grep: giriş noktalarını (entry point) bul (fonksiyon tanımı, export)
2. Read: bulunan dosyaları oku
3. Grep: kullanımları bul (import, çağrılar)
4. Read: tüketici (consumer) dosyaları oku
5. Tam bir resme sahip olana kadar tekrarla
```

## 13.3 Yedek (Fallback): Edit Yerine Read + Write

Edit, benzersiz olmayan bir metin eşleşmesi nedeniyle başarısız olduğunda:
1. Read — tam dosya içeriğini yükle
2. İçeriği programatik olarak değiştir
3. Write — güncellenmiş sürümü yaz

---

# BÖLÜM II: SINAV ALANI (DOMAIN) NOTLARI

---

# Alan 1: Agent Mimarisi ve Orchestration (%27)

## 1.1 Otonom Görev Yürütme için Agentic Loop Tasarımı

### Temel bilgi:
- Agent loop yaşam döngüsü: bir Claude isteği gönder, `stop_reason`'ı kontrol et (`"tool_use"` vs `"end_turn"`), tool'ları çalıştır, bir sonraki iterasyon için sonuçları döndür
- Tool sonuçları, modelin bir sonraki eyleme karar verebilmesi için konuşma geçmişine eklenir
- Model-driven karar verme (Claude bir sonraki tool'u seçer) vs hard-coded karar ağaçları

### Temel beceriler:
- Akış kontrolü: `stop_reason = "tool_use"` olduğunda döngüyü sürdür ve `"end_turn"`'de durdur
- İterasyonlar arasında tool sonuçlarını context'e ekleme
- Kaçınılacak anti-pattern'ler: tamamlanma için assistant metnini parse etme, birincil durdurma mekanizması olarak keyfi iterasyon limitleri kullanma

## 1.2 Multi-agent Sistemleri Orchestrate Etme (Coordinator–Subagent)

### Temel bilgi:
- Hub-and-spoke mimarisi: coordinator, tüm agent'lar arası iletişime, hata yönetimine ve yönlendirmeye sahiptir
- Subagent'ler izole context ile çalışır — coordinator'ın geçmişini otomatik devralmazlar
- Coordinator sorumlulukları: görev ayrıştırma, delegasyon, sonuç toplama, subagent'lerin dinamik seçimi
- Coordinator'ın aşırı dar ayrıştırma yapma riski

### Temel beceriler:
- Tekrarı en aza indirmek için araştırma kapsamını subagent'ler arasında bölme
- İteratif iyileştirme (iterative refinement) döngüleri uygulama (coordinator sentezi değerlendirir ve görevleri yeniden yönlendirir)
- Gözlemlenebilirlik için tüm iletişimi coordinator üzerinden yönlendirme

## 1.3 Subagent Çağrılarını, Context Aktarımını ve Oluşturmayı Yapılandırma

### Temel bilgi:
- `Task` tool'u subagent'leri oluşturur; coordinator'ın `allowedTools`'u `"Task"` içermelidir
- Subagent context'i prompt'ta açıkça dahil edilmelidir; subagent'ler üst (parent) context'i devralmaz
- `AgentDefinition` yapılandırması: açıklamalar, system prompt'lar, tool kısıtlamaları
- Alternatifleri keşfetmek için `fork_session` ile oturum yönetimi

### Temel beceriler:
- Önceki agent'lardan gelen tam çıktıları subagent prompt'una dahil etme
- Context aktarırken veriyi metadata'dan ayırmak için yapılandırılmış formatlar kullanma
- Tek bir coordinator turunda birden çok `Task` çağrısıyla paralel subagent'ler oluşturma
- Coordinator prompt'larını adım adım talimatlar yerine hedefler ve kalite kriterleri cinsinden yazma

## 1.4 Enforcement ve Handoff Desenleriyle Çok Adımlı İş Akışları Uygulama

### Temel bilgi:
- **Programatik enforcement** (hook'lar, ön koşullar) ile bir iş akışını sıralamak için **prompt rehberliği** arasındaki fark
- Deterministik garantilere ihtiyaç duyduğunuzda (örneğin, finansal işlemlerden önce kimlik doğrulaması), tek başına prompt'lar yetersizdir
- Escalation sırasında yapılandırılmış handoff protokolleri (müşteri ID, neden, önerilen eylem)

### Temel beceriler:
- Önceki adımlar tamamlanana kadar sonraki çağrıları engelleyen programatik ön koşullar (örneğin, `get_customer` doğrulanmış bir ID döndürene kadar `process_refund`'u engelle)
- Çok yönlü müşteri taleplerini ayrı kalemlere ayrıştırma
- Bir insana escalate ederken yapılandırılmış özetler üretme

## 1.5 Tool Çağrılarını Yakalamak ve Veriyi Normalleştirmek için Agent SDK Hook'ları

### Temel bilgi:
- Tool sonuçlarını model tüketmeden önce yakalamak için hook desenleri (örneğin `PostToolUse`)
- Uyumluluk kurallarını uygulamak için giden çağrıları yakalayan hook'lar (örneğin, bir eşiğin üzerindeki iadeleri engelle)
- Hook'lar **deterministik garantiler** sağlar; prompt talimatları **olasılıksal uyum** sağlar

### Temel beceriler:
- Veri formatlarını normalleştirmek için `PostToolUse` hook'ları (Unix timestamp'ler, ISO 8601, sayısal durum kodları)
- Politikayı ihlal eden eylemleri escalation'a yönlendirerek engelleyen interception hook'ları
- İş kuralları garantili uyum gerektirdiğinde prompt yerine hook seçme

## 1.6 Karmaşık İş Akışları için Görev Ayrıştırma Stratejileri

### Temel bilgi:
- **Sabit pipeline'lar** (prompt chaining) vs ara sonuçlara dayalı **dinamik uyarlanabilir ayrıştırma**
- Prompt chaining: ardışık adımlar (her dosyayı ayrı ayrı analiz et, sonra bir entegrasyon geçişi çalıştır)
- Keşfedilenlere göre alt görevler üreten uyarlanabilir araştırma planları

### Temel beceriler:
- Öngörülebilir çok yönlü incelemeler için prompt chaining; açık uçlu araştırmalar için dynamic decomposition kullanma
- Büyük kod incelemelerini dosya başına analiz + ayrı bir dosyalar arası entegrasyon geçişine bölme
- Açık uçlu görevleri ayrıştırma: önce yapıyı haritala, sonra önceliklendirilmiş bir plan oluştur

## 1.7 Oturum State'i, Sürdürme (Resuming) ve Forking

### Temel bilgi:
- Adlandırılmış oturumları sürdürmek için `--resume <session-name>`
- Paylaşılan context'ten bağımsız araştırma dalları oluşturmak için `fork_session`
- Oturumları sürdürürken agent'a dosya değişikliklerini bildirmenin önemi
- Yapılandırılmış bir özetle yeni bir oturum, bayatlamış sonuçlarla sürdürmekten daha güvenilir olabilir

### Temel beceriler:
- Adlandırılmış araştırma oturumlarını sürdürmek için `--resume` kullanma
- Yaklaşımları paralel karşılaştırmak için `fork_session` kullanma
- Sürdürme (context hâlâ güncel) vs yeni oturum başlatma (sonuçlar bayat) arasında seçim yapma

---

# Alan 2: Tool Tasarımı ve MCP Entegrasyonu (%18)

## 2.1 Net Açıklamalarla Tool Arayüzleri Tasarlama

### Temel bilgi:
- Tool açıklamaları, bir LLM'in tool seçmek için kullandığı **birincil mekanizmadır**; minimal açıklamalar güvenilmez seçime yol açar
- Girdi formatlarını, örnek sorguları, uç durumları ve uygulanabilirlik sınırlarını dahil etmenin önemi
- Belirsiz veya örtüşen açıklamalar yanlış yönlendirmeye (misrouting) neden olur
- System prompt ifadesi, tool'larla istenmeyen ilişkiler oluşturabilir

### Temel beceriler:
- Her tool'u benzer alternatiflerden net şekilde ayıran açıklamalar yazma
- İşlevsel örtüşmeyi ortadan kaldırmak için tool'ları yeniden adlandırma (örneğin, `analyze_content` -> `extract_web_results`)
- Genel amaçlı tool'ları, net girdi/çıktı sözleşmeleri (contracts) olan özelleşmiş tool'lara bölme

## 2.2 MCP Tool'ları için Yapılandırılmış Hata Yanıtları Uygulama

### Temel bilgi:
- MCP tool yanıtlarındaki `isError` bayrağı
- **Transient hatalar** (timeout'lar), **validation hataları** (kötü girdi), **business hataları** (politika ihlalleri) ve **erişim/permission hataları** arasındaki fark
- Genel hatalar ("Operation failed") doğru kurtarma kararlarını engeller
- Retry edilebilir ve edilemez hatalar arasındaki fark

### Temel beceriler:
- `errorCategory` (transient/validation/permission), `isRetryable` ve okunabilir bir mesaj gibi yapılandırılmış metadata döndürme
- Net kullanıcıya dönük açıklamalarla iş kuralı ihlalleri için `retryable: false` kullanma
- Transient hatalar için subagent'ler içinde yerel kurtarma; yalnızca çözemedikleri hataları iletme
- Erişim hatalarını (retry kararı) geçerli boş sonuçlardan (eşleşme yok) ayırma

## 2.3 Tool'ları Agent'lar Arasında Tahsis Etme ve `tool_choice` Yapılandırma

### Temel bilgi:
- Agent başına çok fazla tool (örneğin, 4–5 yerine 18) tool seçim güvenilirliğini **azaltır**
- Uzmanlık alanı dışında tool'lara sahip agent'lar bunları kötüye kullanma eğilimindedir
- Kapsamlandırılmış (scoped) tool erişimi: yalnızca role-ilgili tool'lar artı sınırlı bir cross-role yardımcı seti
- `tool_choice`: `"auto"`, `"any"` ve zorunlu tool seçimi (`{"type": "tool", "name": "..."}`)

### Temel beceriler:
- Her subagent'in tool setini rolü için ilgili olanla kısıtlama
- Genel tool'ları kısıtlanmış alternatiflerle değiştirme (örneğin, `fetch_url` -> `load_document`)
- Metin yanıtı yerine bir tool çağrısı garantilemek için `tool_choice: "any"` kullanma
- Yürütme sırasını garantilemek için belirli bir tool'u zorlama

## 2.4 MCP Server'larını Claude Code ve Agent İş Akışlarına Entegre Etme

### Temel bilgi:
- MCP server kapsamı: ekipler için proje (`.mcp.json`) vs deneyler için kullanıcı (`~/.claude.json`)
- Secret yönetimi için `.mcp.json`'da environment variable yer değiştirme (örneğin, `${GITHUB_TOKEN}`)
- Bağlı tüm MCP server'larından gelen tool'lar bağlantıda keşfedilir ve aynı anda kullanılabilir
- Keşif amaçlı tool çağrılarını azaltmak için "içerik katalogları" olarak MCP resource'ları (görev özetleri, veritabanı şemaları)

### Temel beceriler:
- Paylaşılan MCP server'larını env-var tabanlı token'larla proje `.mcp.json`'da yapılandırma
- Kişisel/deneysel server'ları `~/.claude.json`'da tutma
- Standart entegrasyonlar için custom server'lar yerine community MCP server'larını tercih etme

## 2.5 Built-in Tool'ları Seçme ve Uygulama (Read, Write, Edit, Bash, Grep, Glob)

### Temel bilgi:
- **Grep**: dosya içeriklerinde arama (fonksiyon adları, hata mesajları, import'lar)
- **Glob**: ad/uzantı desenlerine göre dosya bulma
- **Read/Write**: tam dosya işlemleri; **Edit**: benzersiz metin eşleşmeleriyle kesin değişiklikler
- Edit, benzersiz olmayan eşleşmeler nedeniyle başarısız olursa, Read + Write'a geç

### Temel beceriler:
- İçerik araması için Grep ve desenlere göre dosya keşfi için Glob kullanma
- Anlayışı kademeli oluşturma: giriş noktaları için Grep, sonra akışları izlemek için Read
- Fonksiyon kullanımını wrapper modüller aracılığıyla izleme

---

# Alan 3: Claude Code Yapılandırması ve İş Akışları (%20)

## 3.1 Hiyerarşi, Kapsam ve Modüler Organizasyon ile CLAUDE.md Yapılandırma

### Temel bilgi:
- CLAUDE.md hiyerarşisi: user (`~/.claude/CLAUDE.md`), project (`.claude/CLAUDE.md` veya kök `CLAUDE.md`) ve directory-level (alt dizinlerde CLAUDE.md)
- User-level ayarlar yalnızca tek kullanıcı için geçerlidir ve VCS aracılığıyla paylaşılmaz
- CLAUDE.md'yi modüler hale getirmek için harici dosyalara referans veren `@path` söz dizimi (örneğin, `@./standards/coding-style.md`)
- Monolitik bir CLAUDE.md yerine konu odaklı kural dosyaları için `.claude/rules/` dizini

### Temel beceriler:
- Hiyerarşi sorunlarını teşhis etme (yeni bir ekip üyesi, project-level yerine user-level olduğu için talimatları kaçırır)
- Her paketin CLAUDE.md'sine standartları seçici olarak dahil etmek için `@path` kullanma (örneğin, `@./standards/testing.md`)
- Büyük CLAUDE.md'yi birden çok `.claude/rules/` dosyasına bölme (testing.md, api-conventions.md, deployment.md)

## 3.2 Özel Slash Command'ler ve Skill'ler Oluşturma ve Yapılandırma

### Temel bilgi:
- `.claude/commands/` içindeki **proje command'leri** (VCS ile paylaşılır) vs `~/.claude/commands/` içindeki **kullanıcı command'leri**
- `SKILL.md` frontmatter ile `.claude/skills/` içindeki skill'ler: `context: fork`, `allowed-tools`, `argument-hint`
- `context: fork`, skill'i izole bir subagent context'inde çalıştırır, böylece ana oturumu kirletmez
- Kişisel skill varyantları farklı adlar altında `~/.claude/skills/` içinde bulunabilir

### Temel beceriler:
- Tüm ekibin bunlara sahip olması için proje slash command'lerini `.claude/commands/` içinde saklama
- Ayrıntılı çıktısı olan skill'leri izole etmek için `context: fork` kullanma
- Bir skill'in kullanabileceği tool'ları kısıtlamak için `allowed-tools` kullanma
- Geliştiricilere gerekli parametreleri sormak için `argument-hint` kullanma

## 3.3 Koşullu Konvansiyon Yüklemesi için Path'e Özgü Kurallar Kullanma

### Temel bilgi:
- `.claude/rules/` dosyaları, glob desenlerine göre kuralları etkinleştirmek için YAML frontmatter `paths` içerebilir
- Path-scoped kurallar, yalnızca eşleşen dosyalar düzenlenirken yüklenir ve context ile token tasarrufu sağlar
- Konvansiyonlar birçok dizine uygulandığında (örneğin, testler) glob tabanlı path kuralları directory-level CLAUDE.md'ye tercih edilebilir

### Temel beceriler:
- Yalnızca eşleşen dosyalarda çalışırken yüklemek için `paths: ["terraform/**/*"]` ile `.claude/rules/` dosyaları oluşturma
- Konvansiyonları konumdan bağımsız olarak dosya tipine göre uygulamak için glob desenleri (`**/*.test.tsx`) kullanma
- Konvansiyonlar kod tabanına yayıldığında directory-level CLAUDE.md yerine path'e özgü kuralları tercih etme

## 3.4 Planning Mode ne zaman vs Doğrudan Yürütme ne zaman

### Temel bilgi:
- **Planning mode**: büyük değişiklikler, birden çok uygulanabilir yaklaşım ve mimari kararlar içeren karmaşık görevler için
- **Doğrudan yürütme**: basit, iyi anlaşılmış değişiklikler için (örneğin, tek bir doğrulama ekleme)
- Planning mode, değişiklik yapmadan önce kod tabanının güvenli keşfini sağlar
- Explore subagent, ayrıntılı keşif çıktısını izole eder

### Temel beceriler:
- Mimari sonuçları olan görevler için planning mode kullanma (microservices, 45+ dosyaya dokunan migration'lar)
- Net bir stack trace ve tek dosyalı düzeltmeler için doğrudan yürütme kullanma
- Çok aşamalı görevlerde context window tükenmesini önlemek için Explore subagent kullanma
- Yaklaşımları birleştirme: keşif için planla, sonra uygulama için yürüt

## 3.5 Kademeli İyileştirme (Iterative Refinement) ile Aşamalı Gelişim

### Temel bilgi:
- Somut girdi/çıktı örnekleri, beklentileri iletmenin en etkili yoludur
- **Test-driven iterasyon**: önce testleri yaz, sonra başarısızlıklara göre iterasyon yap
- "Interview" deseni: Claude, açık olmayan tasarım hususlarını ortaya çıkarmak için sorular sorar
- Tüm sorunları tek bir mesajda verme (birbirine bağımlıysa) vs ardışık verme (bağımsızsa)

### Temel beceriler:
- Dönüşüm gereksinimlerini netleştirmek için 2–3 somut girdi/çıktı örneği sağlama
- Uygulamadan önce beklenen davranış, uç durumlar ve performans gereksinimleriyle test setleri oluşturma
- Tasarım yönlerini (cache invalidation, failure mode'ları) ortaya çıkarmak için interview desenini kullanma
- Uç durumlar için örnek girdiler ve beklenen çıktılarla somut test vakaları sağlama

## 3.6 Claude Code'u CI/CD Pipeline'larına Entegre Etme

### Temel bilgi:
- Otomatik pipeline'larda non-interaktif mod için `-p` (veya `--print`) bayrağı
- CI'da yapılandırılmış çıktı için `--output-format json` ve `--json-schema`
- CLAUDE.md, CI tarafından tetiklenen Claude Code için proje context'i sağlar (test standartları, inceleme kriterleri)
- **Oturum context izolasyonu**: kodu üreten aynı oturum, onu incelemekte bağımsız bir instance'tan daha az etkilidir

### Temel beceriler:
- İnteraktif girdide takılmamak için Claude Code'u CI'da `-p` ile çalıştırma
- Yapılandırılmış sonuçlar için `--output-format json` + `--json-schema` kullanma (örneğin, inline PR yorumları)
- Yeni commit'ler sonrasında yeniden çalıştırırken önceki inceleme sonuçlarını dahil etme (yalnızca yeni/düzeltilmemiş sorunları bildir)
- Test üretim kalitesini artırmak için test standartlarını ve mevcut fixture'ları CLAUDE.md'de belgeleme
- Tekrarı önlemek ve stili tutarlı tutmak için yeni testler üretirken mevcut test dosyalarını context'e dahil etme

---

# Alan 4: Prompt Engineering ve Yapılandırılmış (Structured) Çıktı (%20)

## 4.1 Doğruluğu Artırmak için Açık Kriterlerle Prompt Tasarlama

### Temel bilgi:
- Açık kriterler belirsiz talimatlardan daha etkilidir (örneğin, "yorumun doğruluğunu kontrol et" yerine "yalnızca kodla çeliştiğinde yorumları flag'le")
- "be more conservative" gibi genel rehberlik, somut kategorik kriterlerden daha kötü çalışır
- False positive'lerin geliştirici güvenine etkisi: bazı kategorilerdeki yüksek false-positive oranları, doğru kategorilere olan güveni baltalar

### Temel beceriler:
- İnceleme kriterlerini tanımlama: neyin bildirileceği (bug'lar, güvenlik) vs neyin yok sayılacağı (küçük stil)
- Yüksek false-positive oranlı kategorileri geçici olarak devre dışı bırakma
- Her seviye için kod örnekleriyle açık severity kriterleri tanımlama

## 4.2 Çıktı Tutarlılığını Artırmak için Few-shot Prompting Kullanma

### Temel bilgi:
- Few-shot örnekler, tutarlı formatlanmış, eyleme dönük çıktı üretmenin en etkili yöntemidir
- Few-shot, belirsiz durumların ele alınışını gösterebilir (tool seçimi, test kapsamındaki boşluklar)
- Few-shot, modelin yalnızca varsayılanları tekrarlaması yerine yeni desenlere genelleştirmesine yardımcı olur
- Few-shot, çıkarma görevlerinde halüsinasyonları azaltabilir

### Temel beceriler:
- Belirsiz senaryolar için gerekçeyle 2–4 hedefli örnek sağlama
- Çıktı formatını gösteren few-shot örnekler dahil etme (location, issue, severity, suggested fix)
- Kabul edilebilir kod desenlerini gerçek sorunlardan ayıran örnekler sağlama
- Farklı yapılara sahip belgelerden doğru çıkarım örnekleri sağlama

## 4.3 `tool_use` ve JSON Şemalarıyla Yapılandırılmış Çıktıyı Zorlama

### Temel bilgi:
- JSON şemalarıyla `tool_use`, şemaya uygun çıktıyı garanti etmenin ve JSON syntax hatalarını ortadan kaldırmanın en güvenilir yoludur
- `tool_choice: "auto"` ile model metin döndürebilir; `"any"` ile bir tool çağırmalıdır; zorunlu seçim belirli bir tool'u seçer
- Katı JSON şemaları syntax hatalarını ortadan kaldırır ancak anlamsal hataları engellemez (toplamlar tutmaz; değerler yanlış alanlarda)
- Şema tasarımı: required vs optional alanlar; genişletilebilirlik için "other" artı bir detay string'i ile enum'lar

### Temel beceriler:
- JSON şemalarıyla extraction tool'ları tanımlama ve `tool_use` sonuçlarından veriyi parse etme
- Birden çok şema olduğunda yapılandırılmış çıktıyı garantilemek için `tool_choice: "any"` kullanma
- Belirli bir tool çağrısını zorlama: `tool_choice: {"type": "tool", "name": "extract_metadata"}`
- Değer uydurmaktan kaçınmak için kaynak bilgi içermeyebildiğinde alanları optional/nullable yapma
- Genişletilebilir kategorizasyon için `"unclear"` ve `"other"` gibi enum değerleri artı detay alanları kullanma

## 4.4 Çıkarma Kalitesi için Doğrulama, Retry ve Geri Bildirim Döngüleri Uygulama

### Temel bilgi:
- Hata geri bildirimiyle retry: düzeltmeleri yönlendirmek için retry prompt'una somut doğrulama hatalarını dahil etme
- Bilgi kaynakta yoksa retry'lar etkisizdir
- Geri bildirim döngüsü tasarımı: bir bulguyu tetikleyen deseni izleme (`detected_pattern`)
- Anlamsal hatalar (toplamlar uyuşmuyor) vs syntax hataları (`tool_use` ile ele alınır)

### Temel beceriler:
- Orijinal belge, hatalı bir çıkarım ve belirli doğrulama hatalarıyla takip (follow-up) prompt'ları
- Retry'nin etkisiz olacağını belirleme (gerekli bilgi yalnızca harici bir belgede)
- False positive'leri analiz etmek için bulgulara `detected_pattern` alanları dahil etme
- Tutarsızlıkları tespit etmek için hem `calculated_total` hem de `stated_total`'ı çıkararak self-correction tasarlama

## 4.5 Verimli Batch İşleme Stratejileri Tasarlama

### Temel bilgi:
- Message Batches API: %50 tasarruf, 24 saate kadar işleme penceresi, latency SLA garantisi yok
- Batch işleme, engellemeyen (non-blocking) görevler için uygundur (gece raporları, denetimler) ve engelleyen (blocking) görevler için uygun değildir (pre-merge kontroller)
- Batch API, tek bir istek içinde multi-turn tool calling'i desteklemez
- `custom_id` alanları, batch'ler içinde istek/yanıtı ilişkilendirir

### Temel beceriler:
- Engelleyen kontroller için senkron API; gece/haftalık iş yükleri için Batch API kullanma
- Batch gönderim sıklığını SLA ihtiyaçlarına göre planlama (örneğin, 24 saat işleme ile 30 saatlik garanti için 4 saatlik pencereler)
- Yalnızca başarısız belgeleri yeniden göndererek başarısızlıkları ele alma (`custom_id` ile tanımlanır)
- Büyük ölçekli işleme öncesinde bir örnek kullanarak prompt'ları iyileştirme

## 4.6 Multi-instance ve Multi-pass İnceleme Mimarileri Tasarlama

### Temel bilgi:
- Self-review sınırlamaları: model kendi muhakeme context'ini korur ve kendi kararlarını sorgulama olasılığı daha düşüktür
- Bağımsız inceleme instance'ları (üretim context'i olmadan), ince sorunları bulmada daha iyidir
- Multi-pass inceleme: attention dilution'ı önlemek için dosya başına yerel analiz + dosyalar arası entegrasyon geçişi

### Temel beceriler:
- Değişiklikleri üretim context'i olmadan incelemek için ikinci bağımsız bir Claude instance'ı kullanma
- Çok dosyalı incelemeleri dosyalar arası veri akışı analizi için dosya başına geçişler + entegrasyon geçişlerine bölme
- İncelemeleri kalibre edilmiş şekilde yönlendirmek için öz değerlendirme güveniyle doğrulama geçişleri kullanma

---

# Alan 5: Context Yönetimi ve Güvenilirlik (%15)

## 5.1 Kritik Bilgiyi Korumak için Konuşma Context'ini Yönetme

### Temel bilgi:
- Progressive summarization riskleri: sayısal değerler, yüzdeler ve tarihler belirsiz özetlere yoğunlaştırılır
- Lost-in-the-middle etkisi: modeller uzun girdilerin başını ve sonunu güvenilir işler ancak ortadaki bulguları kaçırabilir
- Tool çıktıları, ilgiyle orantısız şekilde context'te birikebilir (5 gerekirken 40+ alan)
- Sonraki API isteklerinde tüm konuşma geçmişini göndermenin önemi

### Temel beceriler:
- İşlemsel olguları, özetlenen geçmişin dışında kalıcı bir "case facts" bloğuna çıkarma
- Ayrıntılı tool çıktılarını ilgili alanlara kırpma
- Kilit bulguları, açık bölüm başlıklarıyla toplanan verinin başına yerleştirme
- Subagent'lerden yapılandırılmış çıktılarda metadata (tarihler, kaynaklar) içermesini isteme

## 5.2 Etkili Escalation Desenleri Tasarlama ve Belirsizliği Çözme

### Temel bilgi:
- Uygun escalation tetikleyicileri: bir insan için açık talep, politika boşlukları/istisnalar, ilerleme kaydedememe
- Anında escalation (açık talep) vs çözüm denemesi (agent kapsamı dahilinde)
- Duygu analizi ve modelin güven öz değerlendirmeleri, vaka karmaşıklığı için güvenilmez proxy'lerdir
- Birden çok müşteri eşleşmesi, sezgisel tahmin değil, ek tanımlayıcılar istemeyi gerektirir

### Temel beceriler:
- System prompt'ta few-shot örneklerle açık escalation kriterleri
- Bir insan için açık talepleri ek araştırma olmadan hemen yürütme
- Politika belirsiz veya belirli bir talep için sessiz olduğunda escalate etme
- Tool sonuçları birden çok eşleşme içerdiğinde ek tanımlayıcılar isteme

## 5.3 Multi-agent Sistemlerde Hata Yayılım (Error Propagation) Stratejileri Uygulama

### Temel bilgi:
- Yapılandırılmış hata context'i (failure tipi, sorgu, kısmi sonuçlar, alternatifler) daha akıllı coordinator kurtarmasını sağlar
- Erişim hatalarını (timeout'lar bir retry kararı gerektirir) geçerli boş sonuçlardan (eşleşme yok) ayırt etme
- Genel hata durumları ("search unavailable") coordinator'dan değerli context'i gizler
- Sessiz bastırma veya tek bir başarısızlıkta tüm iş akışını iptal etme, her ikisi de anti-pattern'dir

### Temel beceriler:
- Yapılandırılmış hata context'i döndürme: failure tipi, ne denendi, kısmi sonuçlar, olası alternatifler
- Erişim hatalarını geçerli boş sonuçlardan ayırma
- Transient hatalar için subagent'lerde yerel kurtarma; yalnızca kurtarılamayan hataları kısmi sonuçlarla iletme
- Sentezde kapsamı annotate etme: neyin iyi desteklendiği vs nerede boşlukların kaldığı

## 5.4 Büyük Kod Tabanlarını Araştırırken Context'i Verimli Yönetme

### Temel bilgi:
- Uzun oturumlarda context bozulması: model, belirli sınıflar yerine "typical patterns"a atıfta bulunarak kararsız yanıtlar üretmeye başlar
- Scratchpad dosyaları, kilit bulguları context sınırları boyunca korur
- Subagent'lere delege etmek, ayrıntılı keşif çıktısını izole eder
- Yapılandırılmış state kalıcılığı crash recovery'yi sağlar

### Temel beceriler:
- Yüksek seviye koordinasyonu ana agent'ta tutarken belirli sorular için subagent'ler oluşturma
- Kilit bulguları saklamak ve daha sonra referans almak için scratchpad dosyaları kullanma
- Sonraki aşama subagent'lerini oluşturmadan önce kilit bulguları özetleme
- Uzun araştırmalar sırasında context kullanımını azaltmak için `/compact` kullanma

## 5.5 İnsan Gözetimi ve Güven Kalibrasyonu ile İş Akışları Tasarlama

### Temel bilgi:
- Toplam metrikler (örneğin, %97 genel doğruluk), belirli belge tiplerinde veya alanlarda zayıf performansı maskeleyebilir
- Tabakalı rastgele örnekleme, yüksek güvenli çıkarımlardaki hata oranlarını ölçer
- Etiketli doğrulama setleri kullanarak alan bazında güven kalibrasyonu
- Otomatikleştirmeden önce doğruluğu belge tipine ve alan segmentine göre doğrulama

### Temel beceriler:
- Yeni hata desenlerini tespit etmek için tabakalı rastgele örnekleme uygulama
- Kararlı performansı doğrulamak için doğruluğu belge tipine ve alana göre analiz etme
- Alan bazında güven skorları üretme ve etiketli veri kullanarak inceleme eşiklerini kalibre etme
- Düşük güvenli veya belirsiz kaynaklı çıkarımları insan incelemesine yönlendirme

## 5.6 Çok Kaynaklı Sentezde Provenance'ı Koruma ve Belirsizliği Ele Alma

### Temel bilgi:
- "claim → source" eşleştirmeleri korunmadan özetleme sırasında atıf kaybolur
- Yapılandırılmış eşleştirmeler, toplama (aggregation) sırasında korunmalıdır
- Çelişkili istatistikleri keyfi olarak bir değer seçmek yerine atıfla annotate ederek ele alma
- Zamansal farklılıkların çelişki olarak yanlış okunmasını önlemek için yayın/toplama tarihlerini dahil etme

### Temel beceriler:
- Subagent'lerden "claim → source" eşleştirmeleri (URL, belge adı, alıntılar) üretmesini isteme
- Raporları, kararlı bulguları tartışmalı olanlardan ayıracak şekilde yapılandırma
- Çelişkili değerleri notlarla koruma ve uzlaştırma (reconciliation) için coordinator'a iletme
- Doğru zamansal yorumlama için yayın tarihlerini dahil etme
- İçeriği tipe göre render etme: finansal veriyi tablo olarak, haberi düz metin olarak, teknik bulguları yapılandırılmış liste olarak

---
