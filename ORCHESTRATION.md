# ORCHESTRATION

## Version

- Status: `v1 released`
- Mode: universal orchestration core

## Purpose

- Этот файл задаёт `v1` ядро оркестрации для проекта.
- Он является source of truth для orchestration behavior, когда оркестрация явно запрошена.
- Главный агент и саб-агенты обязаны опираться на этот файл в orchestration mode.

## How To Use

- Сначала проверь, была ли оркестрация явно запрошена.
- Если нет, не включай orchestration mode и выполняй задачу обычным способом.
- Перед первым orchestration run создай отдельную git-ветку от текущей ветки.
- Имя ветки должно отражать основной смысл orchestration-задачи.
- Если да, собери `orchestration run meta-block`: `run_goal`, `entry_reason`, `scopes`, `subtasks`, `global_verification_requirements`, `merge_policy`, `completion_rule`.
- Перед запуском пройди `pre-run checklist`.
- Во время исполнения следи за `scope boundaries`, `dependencies`, `retry count`, `completeness`, `confidence` и конфликтами.
- Не принимай в merge результаты, которые не прошли `schema valid`, `verification valid` и `merge valid`.
- После завершения пройди `post-run checklist` и оформи финальный output по стандартной схеме.
- Если возникает конфликт, применяй policy `verification -> artifacts -> confidence -> escalation`.
- Если результат неполный, не скрывай это: отражай в `open_issues` и выставляй корректный `final_status`.

## Activation

- Оркестрацию использовать только при явном требовании запроса.
- По умолчанию orchestration mode не включать, чтобы избегать OverPower, over-engineering и лишнего параллелизма.
- Если запрос не требует оркестрации явно, задача выполняется обычным прямым способом.

## Branch Isolation

- Перед запуском оркестратора нужно создать отдельную git-ветку от текущей рабочей ветки.
- Название ветки должно идти по основному смыслу orchestration-задачи.
- Orchestration run не должен стартовать в общей рабочей ветке без branch isolation.
- Цель branch isolation: чистота git history, изоляция многоагентных изменений и предсказуемый handoff результата.

## Task Storage

- В `v1` стек хранения orchestration-задач — `bd`.
- Результат review и planning должен превращаться в `bd` issues, а не в отдельный временный backlog вне трекера.
- Orchestrator task set должен храниться как набор связанных `bd`-задач: epic, child tasks, dependencies, notes, acceptance criteria.
- Review-документы могут быть source artifacts, но не заменяют task tracking.

## Labels For Routing

- Для маршрутизации задач между агентами использовать `bd` labels.
- Рекомендуемый минимальный набор label-префиксов:
- `scope:<name>` — область (`scope:mobile`, `scope:admin`, `scope:desktop`, `scope:cross-frontend`).
- `agent:<name>` — кто должен брать задачу (`agent:orchestrator`, `agent:implementer`, `agent:reviewer`, `agent:docs`, `agent:integration`).
- `mode:<name>` — режим исполнения (`mode:direct`, `mode:orchestrated`).
- `phase:<name>` — стадия (`phase:review`, `phase:plan`, `phase:execute`, `phase:verify`).
- Labels используются для маршрутизации и отбора задач, но не заменяют `dependencies`, `status` и `acceptance criteria`.

## Core Goal

- Обеспечивать предсказуемо высокое качество результата в многообластных задачах за счёт строгой верификации на каждом этапе, единой схемы вывода и контроля согласованности между агентами — при разумной скорости выполнения.

## Priority Order

- Приоритеты ядра: `качество -> скорость -> универсальность`.
- Ключевые акценты: `верификация -> схема -> согласованность`.
- Скорость важна, но не может иметь приоритет над качеством результата.

## Architecture Priorities

### Quality First

- Каждый агент возвращает результат строго по схеме.
- Невалидный по схеме результат не идёт в merge pipeline.
- Перед merge обязателен verification step: полнота, консистентность, отсутствие противоречий.
- При частичном сбое оркестратор делает retry или escalation, а не молчаливую деградацию результата.

### Speed Second

- Параллелизм использовать только там, где он не угрожает качеству.
- Быстрый merge допустим только после прохождения верификации.
- Таймауты мягкие: лучше подождать корректный результат, чем быстро принять плохой.

### Universality Third

- В `v1` строится одно стабильное универсальное ядро.
- Профили и расширения допустимы позже, если не усложняют рабочую логику без пользы.

## V1 Contract

- `v1` фиксирует единое универсальное ядро оркестрации без profile-specific поведения.
- Все правила в этом документе считаются рабочим baseline для orchestration runs в проекте.
- Изменения `v1` должны сохранять совместимость с базовой схемой: `subtask definition`, `verification`, `merge validity`, `final output`.

## Non-Goals

- Не включать оркестрацию по умолчанию.
- Не плодить параллелизм без явной пользы.
- Не подменять архитектурные решения силовым merge.
- Не скрывать конфликты ради красивого результата.
- Не считать скорость важнее верификации.
- Не делать универсальные абстракции ценой усложнения `v1`.
- Не допускать `partial` результат в финальный вывод без явной пометки в `open_issues` и корректного `final_status`.

## When To Orchestrate

- Задача достойна оркестрации, если выполняется хотя бы одно условие:
- затронуты `2+` независимые области;
- есть высокий риск рассогласования между частями задачи, даже если формально это одна область.

## Typical Cases

### Typical YES

- `mobile + admin + desktop`
- `frontend + backend + docs`
- audit по нескольким модулям
- `API + типы + тесты`
- рефакторинг с миграцией БД

### Typical NO

- фикс бага в одном компоненте
- добавление поля в одну форму
- правка текста или стилей в одном месте

## Preventive Use

- Критерий высокого риска рассогласования позволяет включать оркестрацию превентивно, до фактического расхождения.

## Orchestrator Responsibilities

- Определить, нужна ли оркестрация по правилу входа.
- Разбить задачу на независимые подзадачи без пересечений.
- Назначить саб-агентам чёткий `scope`, `expected_output` и критерии проверки.
- Следить, чтобы агенты не дублировали работу.
- Собрать результаты в единую схему.
- Зафиксировать `confidence` и `completeness` каждого саб-агента перед merge.
- Выявить конфликты и рассогласования между ответами.
- Разрешить конфликт по policy либо вынести его явно в итог.
- Провести финальную верификацию и выдать консолидированный результат.

## Standard Subagent Result Schema

- Каждый саб-агент должен возвращать:
- `scope`
- `summary`
- `findings`
- `artifacts`
- `completeness` — `full | partial`
- `confidence` — `high | medium | low`
- `open_questions`
- `blocked_by`
- `handoff_notes`

## Result Schema Rules

- Если обязательные поля отсутствуют, результат невалиден.
- Невалидный результат не попадает в merge pipeline без retry, исправления или escalation.
- `artifacts` должны позволять проверить вывод агента.

## Subagent Result Quality Gate

- Перед merge оркестратор обязан знать по каждому саб-агенту:
- `completeness`
- `confidence`
- Если результат неполный или `confidence = low`, автоматический merge запрещён.

## Merge Validity Levels

### Level 1 - Schema Valid

- Все обязательные поля схемы присутствуют и корректны по формату.

### Level 2 - Verification Valid

- Все назначенные verification steps выполнены.
- Если агенту был выдан checklist проверок, он должен быть отражён в `findings` или `artifacts`.

### Level 3 - Merge Valid

- `completeness = full`
- `confidence = high | medium`
- `blocked_by` пуст
- нет неразрешённых конфликтов с другими агентами

## Merge Rejection Rule

- При провале любого уровня валидности результат нельзя merge-ить.
- Вместо merge оркестратор выполняет retry или escalation.

## Invalid Result Handling

- Отсутствует поле схемы -> retry с уточнением требований.
- `completeness = partial` -> retry или вынесение незавершённости в `open_issues`.
- `confidence = low` -> retry или escalation.
- `blocked_by` непуст -> сначала разрешить зависимость.
- Конфликт с другим агентом -> разрешить или отразить явно в финальном выводе.

## Conflict Resolution Policy

### Step 0 - Diagnosis

- Сначала определить: это реальный конфликт или разный уровень детализации.
- Если различие только в глубине детализации, результаты объединяются.
- Если есть реальное противоречие, применяется resolution chain.

### Step 1 - Verification First

- Приоритет имеет агент, который выполнил обязательные verification steps.

### Step 2 - Artifacts First

- Если оба verification-valid, приоритет у агента с более сильными и конкретными артефактами.
- Файлы, тесты, команды и иные проверяемые доказательства сильнее неподкреплённых утверждений.

### Step 3 - Confidence First

- Если verification и artifacts не определили победителя, приоритет у более высокого `confidence`.
- Порядок силы: `high > medium > low`.

### Step 4 - Escalation

- Если победитель не определён объективно, оркестратор не угадывает.
- Конфликт выносится явно в `conflicts_resolved` или `open_issues`.
- `final_status` в таком случае должен быть `partial` или `blocked`.

## Silent Averaging Ban

- Оркестратор никогда не должен молча усреднять конфликтующие ответы.
- Либо есть объективный победитель, либо конфликт явно отражён в финальном выводе.

## Verification Categories

### Code Verification

- Тесты, типизация, линтер, сборка.
- Обязательна для подзадач, где меняется код.

### Analysis Verification

- Ссылки на файлы, подтверждённые артефакты, сверка нескольких источников.
- Обязательна для исследования, аудита и review-задач.

### Docs Verification

- Проверка синхронизации документации с изменениями.
- Обязательна для изменений, затрагивающих публичный интерфейс, поведение, команды, архитектурные правила или пользовательский flow.

### Integration Verification

- Проверка стыков между областями, модулями или слоями.
- Обязательна для всех многообластных задач.

## Verification Assignment Rules

- Оркестратор назначает только релевантные verification categories.
- Нельзя требовать `code verification` от чисто исследовательского агента; для него обязательна `analysis verification`.
- `integration verification` не может быть опциональной, если задача прошла по правилу входа `2+ области`.
- Каждый назначенный verification step должен быть отражён в `verification_summary` финального вывода.

## Standard Subtask Definition

- Каждая orchestration-подзадача описывается по единой структуре:
- `scope`
- `goal`
- `allowed_actions`
- `forbidden_overlap`
- `expected_output`
- `verification_steps`
- `dependencies`
- `handoff_target`
- `handoff_conditions`

## Standard Verification Step Format

- `type` — `code | analysis | docs | integration`
- `description` — что именно проверить

## Standard Dependency Format

- `agent` — имя или `scope` зависимого агента
- `requires` — какой результат или артефакт нужен

## Communication Model

- В `v1` связь между главным агентом и саб-агентами централизована через оркестратор.
- Оркестратор назначает `scope`, `goal`, `verification_steps`, `dependencies`, `handoff_target` и `handoff_conditions`.
- Саб-агенты не координируются напрямую в свободной форме и не делают peer-to-peer merge.
- Основной канал связи — структурированный результат саб-агента по стандартной схеме.
- Транспорт по умолчанию в `v1` — сообщения внутри orchestration session: оркестратор передаёт подзадачу саб-агенту, саб-агент возвращает один структурированный результат обратно оркестратору.
- Консоль не считается отдельной шиной координации между агентами; она может использоваться только как инструмент выполнения команд внутри своей подзадачи.
- Отдельный файл общей памяти не является обязательным элементом `v1` и не используется как основной канал координации между агентами.
- Если нужна персистентность между циклами, оркестратор может вести run artifact или handoff artifact, но такой файл остаётся журналом оркестратора, а не общей mutable memory для одновременной записи несколькими агентами.
- Если один саб-агент зависит от другого, связь выражается через `dependencies`, а не через неформальный обмен решениями.
- Передача результата идёт через handoff в оркестратор или в явно указанный `handoff_target`.
- Оркестратор остаётся единственной точкой:
- назначения работы;
- проверки валидности результата;
- разрешения конфликтов;
- допуска в merge pipeline;
- формирования финального консолидированного вывода.

## Dependency And Handoff Flow

- Если агент `B` зависит от агента `A`, это фиксируется в `dependencies` как требуемый результат или артефакт.
- Пока зависимость не закрыта, агент `B` не считается merge-ready.
- Если `handoff_conditions` не выполнены, результат не передаётся дальше и агент обязан сигнализировать об этом через `blocked_by` или `open_questions`.
- Любая межагентная согласованность в `v1` проходит через оркестратор, а не через скрытую прямую координацию.

## Shared Memory Rule

- В `v1` запрещено использовать общий редактируемый файл как основную шину коммуникации между саб-агентами.
- Причина: общий mutable state ухудшает трассируемость, создаёт гонки и размывает ответственность за источник истины.
- Если orchestration run требует сохранения состояния между шагами, это состояние должно вестись оркестратором в виде append-only или controlled artifact, а не свободно изменяемой общей памяти.

## Handoff Conditions

- `completeness = full`
- `confidence = high | medium`
- `blocked_by` пуст
- все назначенные `verification_steps` пройдены

## Handoff Rule

- Агент не должен передавать результат, если `handoff_conditions` не выполнены.
- Если условия handoff не выполнены, агент обязан сигнализировать оркестратору через `blocked_by`, `open_questions` или явную пометку незавершённости.

## Retry And Escalation Policy

### Retry Conditions

- Retry допустим только для устранимых проблем.
- Максимум: `2 retry` на одного агента для одной подзадачи.
- Retry делается в случаях:
- неполная схема вывода
- `completeness = partial`
- `confidence = low`
- слабые или отсутствующие артефакты
- verification steps не пройдены, но проблема устранима

### Immediate Escalation Conditions

- Конфликт неразрешим по policy `verification -> artifacts -> confidence`.
- Зависимость внешняя и не может быть снята автоматически.
- Агент дважды вернул невалидный результат и лимит retry исчерпан.

### After Retry Limit

- После исчерпания лимита retry допускается только escalation.
- Если `scope` критичный, `final_status` не может быть `complete`.

## Critical Scope Definition

- `Scope` считается критичным, если выполняется хотя бы одно условие:
- он указан в `dependencies` у других агентов
- он напрямую покрывает основную `goal`
- без него невозможно выполнить `integration verification`

## Failure Decision Tree

- Агент вернул невалидный результат.
- Оркестратор определяет, устранима ли проблема.
- Если проблема устранима, выполняется retry до лимита.
- Если лимит исчерпан или проблема неустранима, выполняется escalation.
- После escalation:
- критичный `scope` ведёт к `final_status = blocked | partial`
- некритичный `scope` фиксируется в `open_issues`, а остальные результаты могут быть merge-нуты

## Orchestration Profiles - V1

- В `v1` работает единое универсальное ядро.
- Профили — справочные кандидаты для `v2`.
- В `v1` оркестратор не выбирает профиль.

### Review Profile Candidate

- Акцент на `analysis verification`
- Агенты работают в read-only режиме
- Конфликты чаще уходят в `open_issues`, чем принудительно разрешаются

### Implementation Profile Candidate

- Обязательны `code verification` и `integration verification`
- Строгий контроль `forbidden_overlap`
- Артефакты — главный критерий качества

### Sync Profile Candidate

- Цель — привести несколько областей к согласованному состоянию
- `integration verification` всегда обязательна
- `final_status = complete` возможен только при полной синхронизации всех `scopes`

## Orchestration Run Meta-Block

- Каждый orchestration run должен быть описан явным meta-block:
- `run_goal`
- `entry_reason`
- `scopes`
- `subtasks`
- `global_verification_requirements`
- `merge_policy`
- `completion_rule`

## Standard Run Scope Format

- `name`
- `critical: true | false`

## Standard Global Verification Requirement Format

- `type` — `code | analysis | docs | integration`
- `description` — что проверяется на уровне всего запуска

## Standard Merge Policy Reference

- Для `v1`: `verification -> artifacts -> confidence -> escalation`

## Completion Rule

- Orchestration run считается завершённым как процесс, когда:
- все критичные `scopes` вернули merge-valid результат или исчерпали retry-лимит с escalation
- все обязательные `global_verification_requirements` выполнены или явно отмечены как невыполненные
- все конфликты либо разрешены, либо явно отражены в `conflicts_resolved` или `open_issues`
- `final_status` определён как `complete | partial | blocked`
- финальный консолидированный output сформирован по стандартной схеме

## Completion Rule Clarification

- `run завершён` — orchestration process полностью отработал: результаты собраны, статус выставлен, handoff оформлен.
- `задача решена` — `final_status = complete`, все `scopes` закрыты и не осталось неразрешённых критичных проблем.
- Run может считаться завершённым даже при `final_status = partial` или `blocked`, если это честный и полностью оформленный итог процесса.

## Final Orchestrator Output Schema

- Финальный output должен содержать:
- `goal`
- `scopes_covered`
- `final_status`
- `key_results`
- `conflicts_resolved`
- `open_issues`
- `verification_summary`
- `artifact_index`
- `next_actions`

## Final Output Field Value

- `verification_summary` даёт прозрачность по качеству.
- `conflicts_resolved` даёт аудит решений оркестратора.
- `open_issues` честно отражает незавершённость.
- `final_status` даёт однозначный сигнал для следующего шага.
- `next_actions` обеспечивает чёткий handoff.

## Operator Checklist

- В `v1` checklist делится на три фазы:
- `pre-run`
- `during-run`
- `post-run`

## Pre-Run Checklist

- Подтверждено, что задача требует оркестрации по правилу входа.
- Создана отдельная git-ветка для orchestration run.
- Определены `run_goal` и `entry_reason`.
- Выделены `scopes` и отмечены критичные области.
- Подзадачи декомпозированы без пересечений.
- Для каждой подзадачи заданы `expected_output`, `verification_steps`, `handoff_conditions`.
- Определены зависимости между агентами.
- Назначены `global_verification_requirements`.
- Зафиксирован `completion_rule`.

## During-Run Checklist

- Каждый агент работает только в пределах своего `scope`.
- Отслеживаются дублирование и overlap.
- Зависимости закрываются в правильном порядке.
- Промежуточные результаты валидируются по схеме.
- `completeness` и `confidence` фиксируются до merge.
- `retry count` не превышает лимит.
- Конфликты выявляются рано.
- Handoff без выполненных `handoff_conditions` запрещён.

## Post-Run Checklist

- Все критичные `scopes` имеют merge-valid результат или корректно escalated.
- Все обязательные `global_verification_requirements` отражены в финальном выводе.
- Конфликты разрешены или явно задокументированы.
- `open_issues` честно отражают незавершённость.
- `final_status` выставлен корректно.
- Финальный output собран по стандартной схеме.
- `artifact_index` позволяет проверить ключевые результаты.
- `next_actions` оформлены как понятный handoff.

## Minimal Template

```yaml
run_goal: "краткая общая цель запуска"

entry_reason: "почему задача подпадает под правило входа оркестрации"

scopes:
  - name: "scope-a"
    critical: true
  - name: "scope-b"
    critical: false

subtasks:
  - scope: "scope-a"
    goal: "что должен получить агент"
    allowed_actions:
      - "разрешённое действие"
    forbidden_overlap:
      - "куда агенту нельзя заходить"
    expected_output:
      - "scope"
      - "summary"
      - "findings"
      - "artifacts"
      - "completeness"
      - "confidence"
      - "open_questions"
      - "blocked_by"
      - "handoff_notes"
    verification_steps:
      - type: analysis
        description: "что проверить"
    dependencies: []
    handoff_target: "orchestrator"
    handoff_conditions:
      - "completeness = full"
      - "confidence = high | medium"
      - "blocked_by пуст"
      - "все verification_steps пройдены"

global_verification_requirements:
  - type: integration
    description: "что проверяется на уровне всего запуска"

merge_policy: "verification -> artifacts -> confidence -> escalation"

completion_rule:
  - "все критичные scopes merge-valid или escalated"
  - "global_verification_requirements выполнены или явно отмечены"
  - "все конфликты разрешены или отражены"
  - "final_status определён"
  - "финальный output сформирован"
```

## Minimal Final Output Template

```yaml
goal: "какая общая задача решалась"

scopes_covered:
  - "scope-a"

final_status: complete | partial | blocked

key_results:
  - "главный итог"

conflicts_resolved:
  - conflict: "описание конфликта"
    resolution: "как разрешён"

open_issues:
  - "что осталось нерешённым"

verification_summary:
  - scope: "scope-a"
    checks_performed: "что проверялось"
    result: "итог проверки"

artifact_index:
  - "файлы, команды, тесты, пути"

next_actions:
  - "что делать дальше и кому"
```

## Canonical Example

- Канонический пример: согласованный review `frontend/mobile`, `frontend/admin` и `frontend/desktop`.
- Пример демонстрирует применение универсального ядра `v1` к многообластной задаче.

### Example Run Meta-Block

```yaml
run_goal: "Провести согласованный review frontend/mobile, frontend/admin и frontend/desktop и собрать единый качественный результат без рассогласования"

entry_reason: "Задача затрагивает 3 независимые области и имеет высокий риск рассогласования критериев review и итоговых рекомендаций"

scopes:
  - name: "frontend/mobile"
    critical: true
  - name: "frontend/admin"
    critical: true
  - name: "frontend/desktop"
    critical: true

subtasks:
  - scope: "frontend/mobile"
    goal: "Проверить mobile frontend и вернуть структурированный список улучшений с привязкой к коду"
    allowed_actions:
      - "читать файлы mobile frontend"
      - "анализировать архитектуру, UX, accessibility, performance, API integration и testing"
    forbidden_overlap:
      - "не анализировать admin frontend"
      - "не анализировать desktop frontend"
      - "не выполнять merge с другими интерфейсами"
    expected_output:
      - "scope"
      - "summary"
      - "findings"
      - "artifacts"
      - "completeness"
      - "confidence"
      - "open_questions"
      - "blocked_by"
      - "handoff_notes"
    verification_steps:
      - type: analysis
        description: "Подтвердить вывод file refs и артефактами из frontend/mobile"
      - type: integration
        description: "Сформулировать findings в формате, совместимом с admin и desktop review"
    dependencies: []
    handoff_target: "orchestrator"
    handoff_conditions:
      - "completeness = full"
      - "confidence = high | medium"
      - "blocked_by пуст"
      - "все verification_steps пройдены"

  - scope: "frontend/admin"
    goal: "Проверить admin frontend и вернуть структурированный список улучшений с привязкой к коду"
    allowed_actions:
      - "читать файлы admin frontend"
      - "анализировать архитектуру, UX, accessibility, performance, API integration и testing"
    forbidden_overlap:
      - "не анализировать mobile frontend"
      - "не анализировать desktop frontend"
      - "не выполнять merge с другими интерфейсами"
    expected_output:
      - "scope"
      - "summary"
      - "findings"
      - "artifacts"
      - "completeness"
      - "confidence"
      - "open_questions"
      - "blocked_by"
      - "handoff_notes"
    verification_steps:
      - type: analysis
        description: "Подтвердить вывод file refs и артефактами из frontend/admin"
      - type: integration
        description: "Сформулировать findings в формате, совместимом с mobile и desktop review"
    dependencies: []
    handoff_target: "orchestrator"
    handoff_conditions:
      - "completeness = full"
      - "confidence = high | medium"
      - "blocked_by пуст"
      - "все verification_steps пройдены"

  - scope: "frontend/desktop"
    goal: "Проверить desktop frontend и вернуть структурированный список улучшений с привязкой к коду"
    allowed_actions:
      - "читать файлы desktop frontend"
      - "анализировать архитектуру, UX, accessibility, performance, API integration и testing"
    forbidden_overlap:
      - "не анализировать mobile frontend"
      - "не анализировать admin frontend"
      - "не выполнять merge с другими интерфейсами"
    expected_output:
      - "scope"
      - "summary"
      - "findings"
      - "artifacts"
      - "completeness"
      - "confidence"
      - "open_questions"
      - "blocked_by"
      - "handoff_notes"
    verification_steps:
      - type: analysis
        description: "Подтвердить вывод file refs и артефактами из frontend/desktop"
      - type: integration
        description: "Сформулировать findings в формате, совместимом с mobile и admin review"
    dependencies: []
    handoff_target: "orchestrator"
    handoff_conditions:
      - "completeness = full"
      - "confidence = high | medium"
      - "blocked_by пуст"
      - "все verification_steps пройдены"

global_verification_requirements:
  - type: analysis
    description: "Каждый интерфейс reviewed с подтверждением findings конкретными file refs"
  - type: integration
    description: "Итоговые findings выровнены по единой схеме и не противоречат друг другу"
  - type: docs
    description: "Консолидированный результат оформлен единообразно по всем трем интерфейсам"

merge_policy: "verification -> artifacts -> confidence -> escalation"

completion_rule:
  - "все критичные scopes merge-valid или escalated"
  - "global_verification_requirements выполнены или явно отмечены"
  - "все конфликты разрешены или отражены"
  - "final_status определён"
  - "финальный output сформирован"
```

### Example Final Output

```yaml
goal: "Провести согласованный review mobile, admin и desktop frontend"

scopes_covered:
  - "frontend/mobile"
  - "frontend/admin"
  - "frontend/desktop"

final_status: complete

key_results:
  - "Для mobile выявлены улучшения в auth/session flow, route-driven navigation и cache update strategy"
  - "Для admin выявлены улучшения в декомпозиции task dialog, auth/session architecture и test coverage"
  - "Для desktop выявлены улучшения в auth/session flow, крупных UI-компонентах и консистентности API integration"
  - "Для всех трех интерфейсов результаты выровнены по единой схеме"

conflicts_resolved:
  - conflict: "Разная детализация между review mobile и desktop по performance-findings"
    resolution: "Не считалось реальным конфликтом; результаты объединены как разная глубина анализа"

open_issues:
  - "Нужна последующая приоритизация cross-frontend backlog по effort и business impact"

verification_summary:
  - scope: "frontend/mobile"
    checks_performed: "analysis verification по file refs и integration verification по единому формату findings"
    result: "passed"
  - scope: "frontend/admin"
    checks_performed: "analysis verification по file refs и integration verification по единому формату findings"
    result: "passed"
  - scope: "frontend/desktop"
    checks_performed: "analysis verification по file refs и integration verification по единому формату findings"
    result: "passed"

artifact_index:
  - "docs/frontend/mobile-issues/2026-03-19-mobile-frontend-review.md"
  - "docs/frontend/admin-issues/2026-03-19-admin-frontend-review.md"
  - "docs/frontend/desktop-issues/2026-03-19-desktop-frontend-review.md"
  - "docs/frontend/README.md"

next_actions:
  - "Собрать единый frontend backlog по трем review"
  - "Отсортировать findings по severity, effort и dependency risk"
  - "Завести связанные bd-задачи для критичных направлений"
```

## Canonical Example Notes

- В этом примере все три `scopes` критичны, потому что они напрямую покрывают `run_goal`.
- `integration verification` обязательна, потому что задача многообластная.
- Разная глубина анализа merge-ится, но реальное противоречие идёт через conflict policy.
- Пример показывает завершённый run с `final_status = complete`, но тот же шаблон допускает честный исход `partial` или `blocked`.
