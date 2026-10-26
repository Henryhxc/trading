# 多 Agent Harness 市場調研與金融產品映射

核查日期：2026-10-25（Asia/Hong_Kong）
觀察窗口：2025-01 至 2026-10。
用途：為 UX Securities 的多 Agent 策略助手原型提供市場與工程背景；不是投資建議，也不是對市場份額、模型效果或商業回報的預測。

## 先說結論

- 近一年可以觀察到 Codex CLI、Claude Code、OpenAI Agents SDK 及 DeepSeek 開放模型在官方倉庫和文件中密集發布或持續演進；但「今年才流行起來」不是由這些一手資料直接證明的市場統計結論。較準確的說法是：**2025–2026 年 Agent 工具、編排框架和開放權重模型的可用性提高，Harness 式工程開始成為可討論的產品組件。**
- Harness 不是另一個模型名稱，而是包住模型的工程層：任務路由、上下文整理、工具／權限、角色協作、結構化輸出、人工核准、評測和運行追蹤。Anthropic 將可重複的編排稱為 workflow，將模型動態決定流程稱為 agent；OpenAI Agents SDK 則把 handoff、tools、guardrails、human-in-the-loop、sessions 和 tracing 列為核心概念。
- 對金融助手，最有價值的不是讓模型直接下單，而是把不同分析視角分工，並把行情來源、時間戳、資料狀態、假設、風險和反例放進可審計的輸出。真實下單工具應維持獨立權限邊界，預設只允許模擬訂單或要求人工確認。
- 本項目目前是交互原型；本研究描述可採用的 Harness 產品設計，不把截圖中的多輪對話宣稱為已完成的多模型並行推理或已驗證的投資能力。

## 1. 市場訊號與時間界線

下表只記錄第一方可核對的公開事實，不把 GitHub star、下載量或新聞熱度當作採用率。

| 時間／項目 | 一手資料可確認的事實 | 對產品研究的含義 |
| --- | --- | --- |
| 2025-01-20 起：DeepSeek-R1 | DeepSeek 官方倉庫介紹 R1-Zero、R1 及蒸餾模型，並公開模型下載入口；README 稱其支援研究社群使用。 | 開放權重讓本地或私有推理成為一條可評估的路徑；不等於資料、訓練配方、商業許可和運維成本全部開放。 |
| 2025-02-22 起：Claude Code | Anthropic 官方倉庫將 Claude Code 定義為位於終端、理解程式庫並執行例行任務和 Git 工作流的 agentic coding tool；倉庫亦列出 plugins，可擴展 custom commands 和 agents。 | 顯示「模型 + 工具 + 工作區 + 指令／插件」已經是可交付的產品形態，金融版本可用相同分層承載分析角色和資料工具。 |
| 2025-03-11 起：OpenAI Agents SDK | OpenAI 官方 SDK README 將 agents、sandbox agents、agents-as-tools／handoffs、tools、guardrails、human-in-the-loop、sessions 和 tracing 列為核心能力。 | 可以把多角色金融分析建模成有權限、可交接、可追蹤的工作流，而不是只做一個聊天框。 |
| 2025-04-13 起：Codex CLI | OpenAI 官方 Codex CLI 倉庫將其定義為在本機運行的 coding agent，並提供本地安裝與文件入口；倉庫以 Apache-2.0 發布。 | Codex 的工程形態可借鑑到「檔案／資料工具、工作區、代理邊界和審批」設計；不能因此推斷金融模型效果。 |
| 2025–2026：DeepSeek-V3.2-Exp 等實驗版本 | DeepSeek 官方 README 將 V3.2-Exp 稱為 experimental release，介紹 Sparse Attention 及長上下文訓練／推理效率探索，並列出 agentic tool use 評測欄位。 | 可把模型效率、長上下文和工具使用納入候選方案評測，但實驗版本不是穩定生產承諾。 |

以上日期是官方 GitHub repository metadata／README 所見的建立或發布時間，不代表這些日期就是市場流行的起點。市場普及率、企業付費部署量和金融領域採用率需要第三方調查或供應商披露，本研究沒有把它們臆測成已知事實。

## 2. Harness 的工程構成

### 2.1 路由、交接和並行

Anthropic 的官方文章《Building effective agents》區分 workflows 與 agents，並列出 prompt chaining、routing、parallelization、orchestrator-workers 和 evaluator-optimizer 等模式。OpenAI Agents SDK 的 handoffs 文件則明確說明：一個 agent 可以把任務交給專門 agent，handoff 會被表示為模型可呼叫的工具，並可帶有輸入過濾、動態啟用和回調。

金融映射可以是：

```text
使用者問題
  ↓ 路由：標的／時間範圍／問題類型／資料新鮮度
  ├─ 宏觀分析師：利率、匯率、選舉與情境假設
  ├─ 產業分析師：產業鏈、競爭格局、催化劑與反例
  ├─ 市場分析師：價格、成交量、波動和回撤
  ├─ 量化分析師：因子、回測口徑、進出場與失效條件
  └─ 風險分析師：集中度、相關性、流動性與最大可承受損失
  ↓
Harness 協調器：核對資料狀態、合併結論、標出分歧和待查證項
```

這是工作流設計，不代表必須由五個獨立模型同時生成；原型可先用固定角色提示、可追蹤的輸入輸出和可替換 provider，等任務評測證明價值後再引入並行模型。

### 2.2 工具權限和資料邊界

OpenAI Agents SDK README 把 tools、guardrails 和 human-in-the-loop 列為能力；其 sandbox agent 文件描述可在容器／工作區中檢查檔案、執行命令、套用 patch 並保留工作區狀態。Claude Code 官方文件和倉庫則持續公開權限、sandbox、hooks、MCP、plugins 及 session 相關能力。

金融產品應採用最小權限：

| 工具類型 | 原型／研究權限 | 交易風險邊界 |
| --- | --- | --- |
| 港股行情與 K 線 | 只讀；回傳 `source`、`data_status`、`as_of` | AkShare 是非官方、延遲／可能過期來源，不能寫成 HKEX 官方即時行情 |
| 基本面／宏觀資料 | 只讀；保留來源、發布時間和版本 | 不能把未核對的新聞或模型猜測當作事實 |
| 回測與情境計算 | 隔離工作區；固定資料切分和費用假設 | 結果必須顯示樣本期、滑點、交易成本和失效條件 |
| 模擬下單 | 可寫入模擬 OMS；使用者確認後執行 | 不能連接真實券商或交易所通道 |
| 真實下單 | 原型不開放 | 未取得券商合同、權限、風控和人工批准前，工具層必須拒絕 |

工具輸入也應限制在結構化 schema；不要讓模型自由拼接 SQL、供應商 URL 或交易指令。所有工具呼叫應帶 request／run／user 關聯標識，日誌去敏，並記錄成功、拒絕、超時和資料過期原因。

### 2.3 上下文、輸出和運行軌跡

OpenAI Agents SDK 把 sessions 用於跨 run 的對話歷史管理，把 tracing 用於查看、除錯和優化 agent workflow。這些概念可轉化為金融助手的上下文合同：

- **使用者上下文**：風險偏好、投資期限、持倉和現金，按使用者隔離；敏感資料不放進提示詞以外的公開日誌。
- **市場上下文**：標的、交易所、幣種、行情來源、`as_of`、資料狀態、時間週期和企業行動口徑。
- **任務上下文**：問題、假設、使用過的工具、輸出 schema、評測版本和人工覆核狀態。
- **結果上下文**：結論、證據欄位、相反情境、風險提示和「資料不足」；對外展示可解釋摘要，不展示或偽造隱藏 chain-of-thought。

最小運行事件可以是：`run_started → route_selected → tool_called → tool_result → role_output → conflict_detected → synthesis_completed → approval_required`。保留軌跡是為了可重現和品質分析，不是為了宣稱模型一定正確。

## 3. 金融多 Agent 的產品機會與限制

### 可測試的產品假設

1. **降低學習門檻**：把宏觀、產業、行情、量化和風險問題拆成可理解的步驟，讓初學者看到「為什麼要查這些資料」，而非只拿到一個買賣結論。
2. **多角度融合**：同一個標的由不同角色提出支持與反對證據，再由協調器整理分歧；不把角色數量當成準確率保證。
3. **左側布局學習**：對 SpaceX／Starlink、半導體週期、選舉事件等案例，呈現催化劑、情境、回撤、連跌天數和失效條件，避免單一敘事驅動下單。
4. **可回顧的策略學習**：把每次問答、資料版本、模型版本和使用者決策記成學習紀錄，讓使用者比較預期與實際結果。

### 不應在原型中暗示的能力

- 多 Agent 的數量不等於獨立證據；如果所有角色共用同一來源或同一錯誤上下文，並行只會放大錯誤。
- 事件案例（例如美國選舉、Tesla 飛行車敘事、SanDisk／Micron／Intel 產業週期）只能展示條件式情境；「會升值」「會下跌」需要明確時間窗、比較基準和資料證據，不能當成確定預測。
- DeepSeek 開放權重不等於免費或零風險。GPU、部署、能耗、並發容量、評測、人力、模型更新和資料合規都要計入總擁有成本；開放模型的權重、程式碼、基座和資料許可也要逐項核對。
- Harness 可以降低工具編排和監控成本，但不會自動消除幻覺、資料延遲、回測偏差、倖存者偏差或金融監管責任。

## 4. 評測和落地建議

先建立金融任務集，再比較托管模型、可合法部署的 DeepSeek 權重和混合架構。每個任務至少保存：問題、日期、標的、可用資料、預期證據、允許工具、答案 schema 和人工評分。

建議指標：

- **事實與來源**：引用欄位是否存在、數字與來源是否一致、行情 `as_of` 和 `data_status` 是否正確。
- **策略可重現性**：同一資料快照是否得到相同計算結果；費用、滑點、回測切分是否顯示。
- **風險完整度**：是否列出反例、最大回撤、流動性、集中度和資料不足。
- **工具安全**：越權讀寫、未批准訂單、跨使用者資料洩露和工具超時是否被阻擋。
- **體驗與成本**：延遲、Token／GPU 成本、並發容量、人工覆核時間，以及初學者完成一個分析任務所需步驟。

落地順序可保持保守：

1. 先做固定角色的只讀分析與結構化輸出。
2. 加入 routing、並行分析、衝突整理和 tracing；所有結果標明資料來源與新鮮度。
3. 加入回測／情境工具和離線評測集，對托管 API 與 DeepSeek 自部署做質量、延遲、成本比較。
4. 僅在模擬 OMS 中開放需確認的模擬訂單；真實 execution adapter 另立項，不能由聊天助手直接取得。

## 5. 一手來源與核查註記

以下均為官方文件、官方 GitHub 倉庫或官方發布頁；連結可隨版本更新，核查時應固定 commit／版本。

- [OpenAI Codex CLI 官方倉庫](https://github.com/openai/codex)：README 對本地 coding agent、安裝方式和 Apache-2.0 的說明；GitHub API metadata 顯示倉庫建立於 2025-04-13。
- [OpenAI Agents SDK 官方倉庫](https://github.com/openai/openai-agents-python)：README 對 agents、tools、handoffs、guardrails、human-in-the-loop、sessions、tracing 和 sandbox agents 的列舉。
- [OpenAI Agents SDK：Handoffs](https://openai.github.io/openai-agents-python/handoffs/)：官方 handoff 概念、工具化交接、輸入過濾及啟用條件。
- [Anthropic Claude Code 官方倉庫](https://github.com/anthropics/claude-code)：README 對 agentic coding tool、終端／IDE／GitHub 使用方式及 plugins 的說明；GitHub API metadata 顯示倉庫建立於 2025-02-22。
- [Anthropic：Building effective agents](https://www.anthropic.com/research/building-effective-agents)：官方對 workflows、agents、routing、parallelization、orchestrator-workers 和 evaluator-optimizer 的定義與取捨。
- [Claude Code 官方文件：Permissions](https://code.claude.com/docs/en/permissions)：工具權限與核准邊界；實作時應按當期文件核對設定語義。
- [DeepSeek-R1 官方倉庫](https://github.com/deepseek-ai/DeepSeek-R1)：R1／R1-Zero、蒸餾模型、下載入口和研究說明；GitHub API metadata 顯示倉庫建立於 2025-01-20。
- [DeepSeek-V3.2-Exp 官方倉庫](https://github.com/deepseek-ai/DeepSeek-V3.2-Exp)：實驗版本、Sparse Attention 和 agentic tool use 評測欄位；實驗版本不代表生產 SLA。

時間和版本限制：本筆記只根據 2026-10-25 可見的一手內容整理。官方倉庫會更新，README 的滾動內容不能代表所有歷史版本；模型許可、基座模型條款、資料權利、服務價格和市場採用率應在部署前重新核查。
