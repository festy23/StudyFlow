# payment-service

## Описание

Сервис отвечает за оплату уроков.

Аутентификацию и авторизацию обеспечивает API Gateway, который прокидывает user_id и user_role в gRPC Context.

---

## Инфа по реализации

- (делаем в последнюю очередь) реализовать механизм ивентов напоминания об оплате:
  - периодически (раз сутки например) запускается воркер
  - получает уроки из schedule_service.ListCompletedUnpaidLessons
  - генерируется ивент-напоминания об оплате и отправляется в кафку

---

## зависимости

- user service
- schedule service
- file service


---

## База данных

![image](db.svg)

### связи с базами данных других сервисов

- lesson_id => schedule_db.lessons.id

---

## Описание gRPC методов

(подробнее со всеми request/response message смотрите в proto файле)

### GetPaymentInfo
**Ошибки:**
- `NOT_FOUND`: урок не найден
- `PERMISSION_DENIED`: не ученик из урока

Получает реквизиты и цену урока по lesson_id. Для получения информации делает запрос в schedule_service.GetLesson и user_service.ResolveTutorStudentContext. Если цены нет в уроке использует стандартную из пары репетитор-ученик.

### SubmitPaymentReceipt
**Ошибки:**
- `INVALID_ARGUMENT`: поля невалидны
- `NOT_FOUND`: урок не найден
- `PERMISSION_DENIED`: не ученик из урока

Создает чек, помечает урок как оплаченный, отправляет ивент-уведомление об оплате урока

### GetReceipt
**Ошибки:**
- `NOT_FOUND`: чек не найден
- `PERMISSION_DENIED`: не участник из урока

Получает чек по id.

### VerifyReceipt
**Ошибки:**
- `NOT_FOUND`: чек не найден
- `PERMISSION_DENIED`: не репетитор из урока

Подтверждение оплаты. Изменяет `is_verified` на true. 

### GetReceiptFile
**Ошибки:**

- `INVALID_ARGUMENT`: receipt_id невалиден

- `NOT_FOUND`: чек или урок не найден

- `PERMISSION_DENIED`: не участник урока

Возвращает временную ссылку на файл чека.
