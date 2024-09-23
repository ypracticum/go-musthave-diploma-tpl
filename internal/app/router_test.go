package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Renal37/go-musthave-diploma-tpl/internal/models"
	mock_models "github.com/Renal37/go-musthave-diploma-tpl/internal/models/mocks"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/services"
	"github.com/Renal37/go-musthave-diploma-tpl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Тестирование маршрута регистрации пользователя
func TestRegisterRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации и JWT
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, nil, nil, nil).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		body            func() io.Reader
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:        "Должен вернуть ошибку валидации из-за отсутствия тела запроса",
			methodName:      "POST",
			targetURL:       "/api/user/register",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Ошибка при разборе данных JSON: unexpected end of JSON input\n",
		},
		{
			testName:   "Должен вернуть ошибку валидации из-за отсутствия логина пользователя",
			methodName: "POST",
			targetURL:  "/api/user/register",
			body: func() io.Reader {
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Запрос не содержит логин или пароль\n",
		},
		{
			testName:   "Должен вернуть ошибку валидации из-за отсутствия пароля пользователя",
			methodName: "POST",
			targetURL:  "/api/user/register",
			body: func() io.Reader {
				Login := "user"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Запрос не содержит логин или пароль\n",
		},
		{
			testName:   "Должен вернуть ошибку, если пользователь уже зарегистрирован",
			methodName: "POST",
			targetURL:  "/api/user/register",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				// Ожидаем вызов GenerateJWT с аргументом "user" и возврат токена без ошибки
				jwtServiceMock.EXPECT().GenerateJWT("user").Return("token", nil)
				// Ожидаем вызов Register с указанными данными и возврат ошибки о уже зарегистрированном пользователе
				authServiceMock.EXPECT().Register(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(services.ErrUserIsAlreadyRegistered)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusConflict,
			expectedMessage: "Пользователь уже зарегистрирован\n",
		},
		{
			testName:   "Должен зарегистрировать пользователя",
			methodName: "POST",
			targetURL:  "/api/user/register",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				// Ожидаем вызов GenerateJWT с аргументом "user" и возврат токена без ошибки
				jwtServiceMock.EXPECT().GenerateJWT("user").Return("token", nil)
				// Ожидаем успешный вызов Register без ошибок
				authServiceMock.EXPECT().Register(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(nil)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "",
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var body io.Reader

			// Если задано тело запроса, получаем его
			if tc.body != nil {
				body = tc.body()
			}

			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "application/json"},
				body,
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

// Тестирование маршрута аутентификации (логина) пользователя
func TestLoginRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации и JWT
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, nil, nil, nil).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		body            func() io.Reader
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
		testHeader      func(t *testing.T, header http.Header)
	}{
		{
			testName:        "Должен вернуть ошибку валидации из-за отсутствия тела запроса",
			methodName:      "POST",
			targetURL:       "/api/user/login",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Ошибка при разборе данных JSON: unexpected end of JSON input\n",
		},
		{
			testName:   "Должен вернуть ошибку валидации из-за отсутствия логина пользователя",
			methodName: "POST",
			targetURL:  "/api/user/login",
			body: func() io.Reader {
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Запрос не содержит логин или пароль\n",
		},
		{
			testName:   "Должен вернуть ошибку валидации из-за отсутствия пароля пользователя",
			methodName: "POST",
			targetURL:  "/api/user/login",
			body: func() io.Reader {
				Login := "user"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Запрос не содержит логин или пароль\n",
		},
		{
			testName:   "Должен вернуть ошибку, если пользователь не существует",
			methodName: "POST",
			targetURL:  "/api/user/login",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				// Ожидаем вызов Login с указанными данными и возврат ошибки о несуществующем пользователе
				authServiceMock.EXPECT().Login(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(services.ErrUserIsNotExist)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Пользователь с логином user не существует\n",
		},
		{
			testName:   "Должен вернуть ошибку, если Неверный пароль",
			methodName: "POST",
			targetURL:  "/api/user/login",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				// Ожидаем вызов Login с указанными данными и возврат ошибки о неверном пароле
				authServiceMock.EXPECT().Login(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(services.ErrPasswordIsIncorrect)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Неверный пароль\n",
		},
		{
			testName:   "Должен вернуть заголовок авторизации",
			methodName: "POST",
			targetURL:  "/api/user/login",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				// Ожидаем вызов GenerateJWT и успешного Login
				jwtServiceMock.EXPECT().GenerateJWT("user").Return("token", nil)
				authServiceMock.EXPECT().Login(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(nil)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "",
			testHeader: func(t *testing.T, header http.Header) {
				// Проверяем, что заголовок Authorization установлен правильно
				assert.Equal(t, "Bearer token", header.Get("Authorization"))
			},
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var body io.Reader

			// Если задано тело запроса, получаем его
			if tc.body != nil {
				body = tc.body()
			}

			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "application/json"},
				body,
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)

			// Если задано тестирование заголовка, выполняем его
			if tc.testHeader != nil {
				tc.testHeader(t, res.Header)
			}
		})
	}
}

// Тестирование маршрута создания заказа
func TestCreateOrderRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации, JWT, заказа и начислений
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)
	accrualServiceMock := mock_models.NewMockAccrualService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, accrualServiceMock, nil).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		body            func() io.Reader
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Должен создать заказ",
			methodName: "POST",
			targetURL:  "/api/user/orders",
			test: func(t *testing.T) {
				// Создаем токен JWT с субъектом "login"
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				// Определяем пользователя
				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				// Ожидаем вызов ValidateToken с токеном "token" и успешную валидацию
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				// Ожидаем вызов VerifyOrderID с "order-id" и возврат true (валидный ID)
				orderServiceMock.EXPECT().VerifyOrderID("order-id").Return(true)
				// Ожидаем вызов CreateOrder с указанными параметрами и успешное создание
				orderServiceMock.EXPECT().CreateOrder(gomock.Any(), "order-id", "user-id").Return(nil)
				// Ожидаем вызов GetUser для получения информации о пользователе
				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				// Ожидаем вызов CalculateAccrual для расчета начислений
				accrualServiceMock.EXPECT().CalculateAccrual("order-id")
			},
			body: func() io.Reader {
				return bytes.NewBuffer([]byte("order-id")) // Тело запроса содержит ID заказа
			},
			expectedCode:    http.StatusAccepted,
			expectedMessage: "",
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var body io.Reader

			// Если задано тело запроса, получаем его
			if tc.body != nil {
				body = tc.body()
			}

			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "text/plain", "Authorization": "Bearer token"},
				body,
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

// Тестирование маршрута получения заказов пользователя
func TestGerOrdersRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации, JWT и заказа
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, nil, nil).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Должен вернуть список заказов",
			methodName: "GET",
			targetURL:  "/api/user/orders",
			test: func(t *testing.T) {
				// Создаем токен JWT с субъектом "login"
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				// Определяем пользователя
				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				// Ожидаем вызов GetUser для получения информации о пользователе
				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				// Ожидаем вызов ValidateToken с токеном "token" и успешную валидацию
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				// Ожидаем вызов GetOrders для получения списка заказов пользователя
				orderServiceMock.EXPECT().GetOrders(gomock.Any(), "user-id").Return([]models.Order{
					{
						ID:         "order-id",
						Status:     "StatusNew",
						Accrual:    nil,
						UploadedAt: utils.RFC3339Date{Time: time.Date(2009, 11, 17, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: `[{"number":"order-id","status":"StatusNew","uploaded_at":"2009-11-17T00:00:00Z"}]`,
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Authorization": "Bearer token"},
				nil, // GET-запрос без тела
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

// Тестирование маршрута получения баланса пользователя
func TestGerBalanceRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации, JWT и баланса
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	balanceServiceMock := mock_models.NewMockBalanceService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, nil, nil, balanceServiceMock).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Должен вернуть баланс пользователя",
			methodName: "GET",
			targetURL:  "/api/user/balance",
			test: func(t *testing.T) {
				// Создаем токен JWT с субъектом "login"
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				// Определяем пользователя
				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				// Ожидаем вызов GetUser для получения информации о пользователе
				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				// Ожидаем вызов ValidateToken с токеном "token" и успешную валидацию
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				// Ожидаем вызов GetUserBalance для получения баланса пользователя
				balanceServiceMock.EXPECT().GetUserBalance(gomock.Any(), "user-id").Return(models.Balance{Current: 100.2, Withdrawn: 100.3}, nil)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: `{"current":100.2,"withdrawn":100.3}`,
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Authorization": "Bearer token"},
				nil, // GET-запрос без тела
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

// Тестирование маршрута создания вывода средств (withdrawal)
func TestCreateWithdrawalRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации, JWT, заказа и баланса
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)
	balanceServiceMock := mock_models.NewMockBalanceService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, nil, balanceServiceMock).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		body            func() io.Reader
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Должен создать вывод средств",
			methodName: "POST",
			targetURL:  "/api/user/balance/withdraw",
			test: func(t *testing.T) {
				// Создаем токен JWT с субъектом "login"
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				// Определяем пользователя и параметры вывода
				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}
				orderID := "withdraw-id"
				sum := 50.2

				// Ожидаем вызов GetUser для получения информации о пользователе
				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				// Ожидаем вызов ValidateToken с токеном "token" и успешную валидацию
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				// Ожидаем вызов VerifyOrderID для проверки валидности ID заказа
				orderServiceMock.EXPECT().VerifyOrderID(orderID).Return(true)
				// Ожидаем вызов GetUserBalance для получения текущего баланса
				balanceServiceMock.EXPECT().GetUserBalance(gomock.Any(), "user-id").Return(models.Balance{Current: 100.2, Withdrawn: 100.3}, nil)
				// Ожидаем вызов CreateWithdrawal для создания вывода средств
				balanceServiceMock.EXPECT().CreateWithdrawal(gomock.Any(), orderID, "user-id", sum).Return(nil)
			},
			body: func() io.Reader {
				ID := "withdraw-id"
				Sum := 50.2

				// Формируем тело запроса с данными вывода
				data, _ := json.Marshal(models.Withdrawal{ID: &ID, Sum: &Sum})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "",
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
				tc.body(),
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

// Тестирование маршрута получения выводов средств пользователя
func TestGetWithdrawalsRoute(t *testing.T) {
	ctrl := gomock.NewController(t) // Создаем контроллер для моков
	defer ctrl.Finish()             // Завершаем контроллер в конце теста

	// Создаем моки для сервисов аутентификации, JWT, заказа и баланса
	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)
	balanceServiceMock := mock_models.NewMockBalanceService(ctrl)

	// Создаем тестовый сервер с маршрутизатором
	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, nil, balanceServiceMock).get(),
	)
	defer testServer.Close() // Закрываем сервер после тестов

	// Определяем набор тестовых случаев
	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Должен вернуть список выводов средств",
			methodName: "GET",
			targetURL:  "/api/user/withdrawals",
			test: func(t *testing.T) {
				// Создаем токен JWT с субъектом "login"
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				// Определяем пользователя
				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				// Ожидаем вызов GetUser для получения информации о пользователе
				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				// Ожидаем вызов ValidateToken с токеном "token" и успешную валидацию
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				// Ожидаем вызов GetWithdrawalFlow для получения списка выводов средств
				balanceServiceMock.EXPECT().GetWithdrawalFlow(gomock.Any(), "user-id").Return([]models.WithdrawalFlowItem{
					{
						OrderID:     "order-id",
						Sum:         123.123,
						ProcessedAt: utils.RFC3339Date{Time: time.Date(2009, 11, 17, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: `[{"order":"order-id","sum":123.123,"processed_at":"2009-11-17T00:00:00Z"}]`,
		},
	}

	// Запуск каждого тестового случая
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Если задано дополнительное тестирование, выполняем его
			if tc.test != nil {
				tc.test(t)
			}

			// Выполняем тестовый запрос
			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Authorization": "Bearer token"},
				nil, // GET-запрос без тела
			)
			res.Body.Close() // Закрываем тело ответа

			// Проверяем ожидаемый статус код
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			// Проверяем ожидаемое сообщение
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}
