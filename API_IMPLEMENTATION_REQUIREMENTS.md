# API Implementation Requirements для Floxy Library

## Анализ нереализованных эндпоинтов

### Реализованные эндпоинты в floxy library:
```
GET /api/workflows                    ✅ Реализован
GET /api/workflows/{id}               ✅ Реализован  
GET /api/workflows/{id}/instances     ✅ Реализован
GET /api/instances                    ✅ Реализован
GET /api/instances/{id}               ✅ Реализован
GET /api/instances/{id}/steps         ✅ Реализован
GET /api/instances/{id}/events        ✅ Реализован
GET /api/stats                        ✅ Реализован
```

### Нереализованные эндпоинты (требуют реализации):

#### 1. `GET /api/stats/summary` 
**Используется в:** `Dashboard.tsx:38`
**Описание:** Возвращает сводную статистику для дашборда
**Ожидаемый формат ответа:**
```typescript
interface SummaryStats {
  total_workflows: number;
  active_instances: number;
  completed_instances: number;
  failed_instances: number;
  // другие поля статистики
}
```

#### 2. `GET /api/instances/active`
**Используется в:** `Dashboard.tsx:39`, `Instances.tsx:27`
**Описание:** Возвращает список активных экземпляров workflow
**Ожидаемый формат ответа:**
```typescript
interface ActiveWorkflow {
  id: string;
  workflow_id: string;
  workflow_name: string;
  status: string;
  started_at: string;
  updated_at: string;
  current_step: string;
  total_steps: number;
  completed_steps: number;
  rolled_back_steps: number;
}
```

## Детальный анализ по компонентам

### Dashboard.tsx
- **Эндпоинт:** `GET /api/stats/summary` ❌ НЕ РЕАЛИЗОВАН
- **Эндпоинт:** `GET /api/instances/active` ❌ НЕ РЕАЛИЗОВАН

### Stats.tsx  
- **Эндпоинт:** `GET /api/stats` ✅ РЕАЛИЗОВАН

### Instances.tsx
- **Эндпоинт:** `GET /api/instances/active` ❌ НЕ РЕАЛИЗОВАН

### Workflows.tsx
- **Эндпоинт:** `GET /api/workflows` ✅ РЕАЛИЗОВАН

### WorkflowDetail.tsx
- **Эндпоинт:** `GET /api/workflows/{id}` ✅ РЕАЛИЗОВАН
- **Эндпоинт:** `GET /api/workflows/{id}/instances` ✅ РЕАЛИЗОВАН

### InstanceDetail.tsx
- **Эндпоинт:** `GET /api/instances/{id}` ✅ РЕАЛИЗОВАН
- **Эндпоинт:** `GET /api/instances/{id}/steps` ✅ РЕАЛИЗОВАН
- **Эндпоинт:** `GET /api/instances/{id}/events` ✅ РЕАЛИЗОВАН
- **Эндпоинт:** `GET /api/workflows/{id}` ✅ РЕАЛИЗОВАН

## Требования к реализации

### 1. Эндпоинт `GET /api/stats/summary`

**Функция:** `handleGetSummaryStats`
**Описание:** Возвращает сводную статистику для главной страницы дашборда

**Ожидаемые поля ответа:**
```go
type SummaryStats struct {
    TotalWorkflows    int `json:"total_workflows"`
    ActiveInstances   int `json:"active_instances"`
    CompletedInstances int `json:"completed_instances"`
    FailedInstances   int `json:"failed_instances"`
    SuccessRate       float64 `json:"success_rate"`
    AvgDuration       float64 `json:"avg_duration_seconds"`
}
```

**SQL запрос (пример):**
```sql
SELECT 
    COUNT(DISTINCT w.id) as total_workflows,
    COUNT(CASE WHEN wi.status = 'running' THEN 1 END) as active_instances,
    COUNT(CASE WHEN wi.status = 'completed' THEN 1 END) as completed_instances,
    COUNT(CASE WHEN wi.status = 'failed' THEN 1 END) as failed_instances,
    AVG(CASE WHEN wi.status = 'completed' THEN EXTRACT(EPOCH FROM (wi.updated_at - wi.created_at)) END) as avg_duration_seconds
FROM workflows w
LEFT JOIN workflow_instances wi ON w.id = wi.workflow_id
```

### 2. Эндпоинт `GET /api/instances/active`

**Функция:** `handleGetActiveInstances`
**Описание:** Возвращает список всех активных экземпляров workflow

**Ожидаемые поля ответа:**
```go
type ActiveWorkflow struct {
    ID              string    `json:"id"`
    WorkflowID      string    `json:"workflow_id"`
    WorkflowName    string    `json:"workflow_name"`
    Status          string    `json:"status"`
    StartedAt       time.Time `json:"started_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    CurrentStep     string    `json:"current_step"`
    TotalSteps      int       `json:"total_steps"`
    CompletedSteps  int       `json:"completed_steps"`
    RolledBackSteps int       `json:"rolled_back_steps"`
}
```

**SQL запрос (пример):**
```sql
SELECT 
    wi.id,
    wi.workflow_id,
    w.name as workflow_name,
    wi.status,
    wi.created_at as started_at,
    wi.updated_at,
    wi.current_step,
    -- Подсчет общего количества шагов из definition
    -- Подсчет завершенных шагов
    -- Подсчет откаченных шагов
FROM workflow_instances wi
JOIN workflows w ON wi.workflow_id = w.id
WHERE wi.status IN ('running', 'pending', 'paused')
ORDER BY wi.created_at DESC
```

## Приоритет реализации

1. **Высокий приоритет:** `GET /api/instances/active` - используется в двух компонентах
2. **Высокий приоритет:** `GET /api/stats/summary` - критичен для дашборда

## Дополнительные замечания

- Все эндпоинты должны возвращать JSON
- Необходима обработка ошибок (404, 500)
- Рекомендуется добавить пагинацию для больших списков
- Стоит рассмотреть кэширование для статистических данных
- Время ответа должно быть оптимизировано для дашборда
