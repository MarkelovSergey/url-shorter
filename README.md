# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Бенчмарки и профилирование памяти

### Запуск бенчмарков

```bash
# Бенчмарки сервиса URL shortener
go test -bench=. -benchmem ./internal/service/urlshorterservice/

# Бенчмарки memory storage
go test -bench=. -benchmem ./internal/storage/memorystorage/

# Бенчмарки handler
go test -bench=. -benchmem ./internal/handler/

# Бенчмарки репозитория
go test -bench=. -benchmem ./internal/repository/urlshorterrepository/

# Профилирование с созданием профиля памяти
PROFILE_NAME=profiles/base.pprof go test -v ./profiles/ -run=TestProfileMemory
```

### Анализ профиля памяти

```bash
# Просмотр топ аллокаторов
go tool pprof -top profiles/base.pprof

# Сравнение профилей
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

### Результаты оптимизации памяти

Результат сравнения профилей памяти (`pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`):

```
File: profiles.test
Type: inuse_space
Showing nodes accounting for 1kB, 0.065% of 1538kB total
Dropped 1 node (cum <= 7.69kB)
      flat  flat%   sum%        cum   cum%
     513kB 33.35% 33.35%      513kB 33.35%  runtime.allocm
    -512kB 33.29% 0.065%     -512kB 33.29%  os.newFile
         0     0% 0.065%     -512kB 33.29%  main.main
         0     0% 0.065%     -512kB 33.29%  os.Create (inline)
         0     0% 0.065%     -512kB 33.29%  os.OpenFile
         0     0% 0.065%     -512kB 33.29%  os.openFileNolog
         0     0% 0.065%      513kB 33.35%  runtime.mcall
         0     0% 0.065%      513kB 33.35%  runtime.newm
         0     0% 0.065%      513kB 33.35%  runtime.park_m
         0     0% 0.065%      513kB 33.35%  runtime.resetspinning
         0     0% 0.065%      513kB 33.35%  runtime.schedule
         0     0% 0.065%      513kB 33.35%  runtime.startm
         0     0% 0.065%      513kB 33.35%  runtime.wakep
         0     0% 0.065%     -512kB 33.29%  testing.(*M).Run
         0     0% 0.065%     -512kB 33.29%  testing.(*M).before
```

**Отрицательные значения (-512kB) показывают, что использование памяти уменьшилось.**

### Выполненные оптимизации

1. **`generateRandomShortCode`** - добавлен `sync.Pool` для переиспользования буферов при генерации коротких кодов
2. **`FindByUserID`** - добавлен предварительный подсчёт для выделения slice с правильной capacity (уменьшение реаллокаций)
3. **`DeleteBatch`** - удалена неиспользуемая аллокация map

### Результаты бенчмарков после оптимизации

```
BenchmarkMemoryStorageDeleteBatch/delete_10_from_1000-16    2652802    405.0 ns/op    0 B/op    0 allocs/op
BenchmarkMemoryStorageDeleteBatch/delete_100_from_10000-16   367806   3276 ns/op      0 B/op    0 allocs/op
BenchmarkMemoryStorageFindByUserID/1000_records_10_users-16   74890  16061 ns/op   8192 B/op    1 allocs/op
BenchmarkMemoryStorageFindByUserID/10000_records_100_users-16 17445  60016 ns/op   8192 B/op    1 allocs/op
```

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**
