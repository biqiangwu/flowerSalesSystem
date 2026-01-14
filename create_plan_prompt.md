**Prompt 1: 生成技术方案**

`@./specs/001-core-functionality/spec.md`

你现在是`flowerSalesSystem`项目的首席架构师。你的任务是基于我提供的`spec.md`以及我们已有的`constitution.md`（你已自动加载），为项目生成一份详细的技术实现方案（`plan.md`）。

**技术栈约束 (必须遵循):**

- **语言**: Go (>=1.25.0)
- **Web框架**: 仅使用标准库 `net/http`，不引入Gin或Echo等外部框架（遵循“简单性原则”）。

**方案内容要求 (必须包含):**

1.  **技术上下文总结:** 明确上述技术选型。
2.  **“合宪性”审查:** 逐条对照`constitution.md`的原则，检查并确认本技术方案符合所有条款（特别是包内聚、错误处理、TDD）。
3.  **项目结构细化:** 明确前端、server、数据库等包的具体职责和依赖关系。
4.  **核心数据结构:** 定义在模块间流转的核心Go `struct`。
5.  **接口设计:** 定义`internal`包对外暴露的关键Interface。

请严格按照`@./.claude/templates/plan-template.md`的模板格式来组织你的输出（如果模板不存在，请自行设计一个结构清晰的Markdown格式）。

完成后，将生成的`plan.md`内容写入到`./specs/001-core-functionality/plan.md`文件中。
