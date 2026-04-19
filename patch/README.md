# patch-ish

Патч для бинарника `client-ish`, предназначенного для запуска в [iSH](https://ish.app/) на iOS.

## Проблема

Go ≥ 1.19 на `linux/386` использует системный вызов **422 (`futex_time64`)** вместо
стандартного **240 (`futex`)**. iSH не поддерживает `futex_time64` и возвращает
`ENOSYS` / `exec format error`, из-за чего бинарник вылетает сразу при старте.

Патч находит инструкцию `MOV EAX, 422` с характерным сигнатурным байтом (`8b 5c 24 04`)
и заменяет только номер вызова на `240`, оставляя остальной код нетронутым.

## Применение

После каждой пересборки `client-ish` запусти патч:

```bash
python3 patch/patch-ish.py client-ish
```

Или с явным путём:

```bash
python3 patch/patch-ish.py /path/to/client-ish
```

## Сборка + патч (полный цикл)

```bash
# 1. Собрать бинарник для iSH
GOOS=linux GOARCH=386 GO386=softfloat CGO_ENABLED=0 \
  go build -o client-ish ./client/

# 2. Запустить патч
python3 patch/patch-ish.py client-ish

# 3. Скопировать в iSH на iPhone
#    (вставить через Files.app / AltStore / SSH и т.д.)
cp client-ish /path/to/iSH/root/client-ish
```

## Технические детали

| Параметр | Значение |
|---|---|
| Искомый паттерн | `b8 a6 01 00 00 8b 5c 24 04` |
| Замена (первые 5 байт) | `b8 f0 00 00 00` |
| Syscall до патча | 422 = `futex_time64` |
| Syscall после патча | 240 = `futex` |

> **Важно:** патч не трогает инструкцию `MOV EBX, [ESP+4]` (байты `8b 5c 24 04`),
> которая читает аргумент `uaddr` со стека. Замена этого поля сломала бы GC Go.
