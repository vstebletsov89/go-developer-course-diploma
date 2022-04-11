package main

//
//Main flow:
//
//read config (RUN_ADDRESS , DATABASE_URI, ACCRUAL_SYSTEM_ADDRESS )
//create repositories + migration (создать необходимые таблицы в БД)
//create service
//create worker pool
//start worker pool
//create router/controller
//start server
//
//
//1) слой контроллер (обработка всех ендпоинтов по ТЗ)
//
//POST /api/user/register — регистрация пользователя; -> RegisterUser
//POST /api/user/login — аутентификация пользователя; -> LoginUser
//POST /api/user/orders — загрузка пользователем номера заказа для расчёта; -> UploadOrder
//GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях; -> GetOrders
//GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя; -> GetBalance
//POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа; -> WithdrawBalance
//GET /api/user/balance/withdrawals — получение информации о выводе средств с накопительного счёта пользователем. -> GetWithdrawals
//
//controller - only parsing data!!!
//
//2) слой репозиторий отвечает за взаимодействие с БД:
//
//UserRepo - RegisterUser/LoginUser - связанная таблица users
//OrderRepo - UploadOrder/GetOrders - связанная таблица orders
//WithDrawRepo - WithdrawBalance/GetWithdrawals - связанная таблица withdraw
//BalanceRepo - SetBalance/GetBalance - связанная таблица balance
//
//
//3) слой сервис:
//
//обрабатывает все совершенные заказы через worker pool и общается с внешней системой accrual для начисления баллов+
//когда совершен заказ создаем worker, чтобы позднее его обработать и начислить баллы используя внешнюю систему и записать в БД (если accrual > 0)
//BalanceRepo -> SetBalance(current balance + accrual) для текущего user
//
//4) Модели:
//
//User(userID, password)
//Order(orderID, userID, status, accrual, uploaded_at)
//Withdraw(orderID, sum, processed_at)
//Balance(userID, current, withdrawn)
//
//Вопросы:
//
//1) как описывать интерфейс взаимодействия с внешним сервисом accrual - отдельный клиент
//2) можно ли использовать jwtauth для аутентификации+
//3) логгер - logrus+ (add log_level to config)
//4) ошибки лучше описать file service/repository (пример ErrTooManyRequests = errors.New("Too Many Requests")); controller no need
func main() {}
