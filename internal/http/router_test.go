package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	mock_models "github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models/mocks"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/services"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, nil, nil, nil).get(),
	)
	defer testServer.Close()

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
			testName:        "Should return a validation error due to missing body",
			methodName:      "POST",
			targetURL:       "/api/user/register",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Error occurred during unmarshaling data unexpected end of JSON input\n",
		},
		{
			testName:   "Should return a validation error due to missing user login",
			methodName: "POST",
			targetURL:  "/api/user/register",
			body: func() io.Reader {
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Request doesn't contain login or password\n",
		},
		{
			testName:   "Should return a validation error due to missing user password",
			methodName: "POST",
			targetURL:  "/api/user/register",
			body: func() io.Reader {
				Login := "user"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Request doesn't contain login or password\n",
		},
		{
			testName:   "Should return error when user is already registered",
			methodName: "POST",
			targetURL:  "/api/user/register",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				jwtServiceMock.EXPECT().GenerateJWT("user").Return("token", nil)
				authServiceMock.EXPECT().Register(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(services.ErrUserIsAlreadyRegistered)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusConflict,
			expectedMessage: "User is already registered\n",
		},
		{
			testName:   "Should register user",
			methodName: "POST",
			targetURL:  "/api/user/register",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				jwtServiceMock.EXPECT().GenerateJWT("user").Return("token", nil)
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

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var body io.Reader

			if tc.body != nil {
				body = tc.body()
			}

			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "application/json"},
				body,
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

func TestLoginRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, nil, nil, nil).get(),
	)
	defer testServer.Close()

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
			testName:        "Should return a validation error due to missing body",
			methodName:      "POST",
			targetURL:       "/api/user/login",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Error occurred during unmarshaling data unexpected end of JSON input\n",
		},
		{
			testName:   "Should return a validation error due to missing user login",
			methodName: "POST",
			targetURL:  "/api/user/login",
			body: func() io.Reader {
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Request doesn't contain login or password\n",
		},
		{
			testName:   "Should return a validation error due to missing user password",
			methodName: "POST",
			targetURL:  "/api/user/login",
			body: func() io.Reader {
				Login := "user"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Request doesn't contain login or password\n",
		},
		{
			testName:   "Should return error when user login isn't exist",
			methodName: "POST",
			targetURL:  "/api/user/login",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				authServiceMock.EXPECT().Login(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(services.ErrUserIsNotExist)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Login user is not exist\n",
		},
		{
			testName:   "Should return error when password isn't correct",
			methodName: "POST",
			targetURL:  "/api/user/login",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

				authServiceMock.EXPECT().Login(gomock.Any(), models.UnknownUser{Login: &Login, Password: &Password}).Return(services.ErrPasswordIsIncorrect)
			},
			body: func() io.Reader {
				Login := "user"
				Password := "123"
				data, _ := json.Marshal(models.UnknownUser{Login: &Login, Password: &Password})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusUnauthorized,
			expectedMessage: "Password is not correct\n",
		},
		{
			testName:   "Should return authorization header",
			methodName: "POST",
			targetURL:  "/api/user/login",
			test: func(t *testing.T) {
				Login := "user"
				Password := "123"

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
				assert.Equal(t, "Bearer token", header.Get("Authorization"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var body io.Reader

			if tc.body != nil {
				body = tc.body()
			}

			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "application/json"},
				body,
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)

			if tc.testHeader != nil {
				tc.testHeader(t, res.Header)
			}
		})
	}
}

func TestCreateOrderRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)
	accrualServiceMock := mock_models.NewMockAccrualService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, accrualServiceMock, nil).get(),
	)
	defer testServer.Close()

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
			testName:   "Should create order",
			methodName: "POST",
			targetURL:  "/api/user/orders",
			test: func(t *testing.T) {
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				orderServiceMock.EXPECT().VerifyOrderID("order-id").Return(true)
				orderServiceMock.EXPECT().CreateOrder(gomock.Any(), "order-id", "user-id").Return(nil)
				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				accrualServiceMock.EXPECT().CalculateAccrual("order-id")
			},
			body: func() io.Reader {
				return bytes.NewBuffer([]byte("order-id"))
			},
			expectedCode:    http.StatusAccepted,
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var body io.Reader

			if tc.body != nil {
				body = tc.body()
			}

			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "text/plain", "Authorization": "Bearer token"},
				body,
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

func TestGerOrdersRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, nil, nil).get(),
	)
	defer testServer.Close()

	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Should return orders",
			methodName: "GET",
			targetURL:  "/api/user/orders",
			test: func(t *testing.T) {
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
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
			expectedMessage: "[{\"number\":\"order-id\",\"status\":\"StatusNew\",\"uploaded_at\":\"2009-11-17T00:00:00Z\"}]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Authorization": "Bearer token"},
				nil,
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

func TestGerBalanceRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	balanceServiceMock := mock_models.NewMockBalanceService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, nil, nil, balanceServiceMock).get(),
	)
	defer testServer.Close()

	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Should return balance",
			methodName: "GET",
			targetURL:  "/api/user/balance",
			test: func(t *testing.T) {
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				balanceServiceMock.EXPECT().GetUserBalance(gomock.Any(), "user-id").Return(models.Balance{Current: 100.2, Withdrawn: 100.3}, nil)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "{\"current\":100.2,\"withdrawn\":100.3}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Authorization": "Bearer token"},
				nil,
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

func TestCreateWithdrawalRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)
	balanceServiceMock := mock_models.NewMockBalanceService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, nil, balanceServiceMock).get(),
	)
	defer testServer.Close()

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
			testName:   "Should create withdraw",
			methodName: "POST",
			targetURL:  "/api/user/balance/withdraw",
			test: func(t *testing.T) {
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}
				orderID := "withdraw-id"
				sum := 50.2

				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				orderServiceMock.EXPECT().VerifyOrderID(orderID).Return(true)
				balanceServiceMock.EXPECT().GetUserBalance(gomock.Any(), "user-id").Return(models.Balance{Current: 100.2, Withdrawn: 100.3}, nil)
				balanceServiceMock.EXPECT().CreateWithdrawal(gomock.Any(), orderID, "user-id", sum).Return(nil)
			},
			body: func() io.Reader {
				ID := "withdraw-id"
				Sum := 50.2

				data, _ := json.Marshal(models.Withdrawal{ID: &ID, Sum: &Sum})
				return bytes.NewBuffer(data)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
				tc.body(),
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}

func TestGetWithdrawalsRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authServiceMock := mock_models.NewMockAuthService(ctrl)
	jwtServiceMock := mock_models.NewMockJWTService(ctrl)
	orderServiceMock := mock_models.NewMockOrderService(ctrl)
	balanceServiceMock := mock_models.NewMockBalanceService(ctrl)

	testServer := httptest.NewServer(
		New(Config{}, authServiceMock, jwtServiceMock, orderServiceMock, nil, balanceServiceMock).get(),
	)
	defer testServer.Close()

	testCases := []struct {
		testName        string
		methodName      string
		targetURL       string
		test            func(t *testing.T)
		expectedCode    int
		expectedMessage string
	}{
		{
			testName:   "Should returns withdrawals",
			methodName: "GET",
			targetURL:  "/api/user/withdrawals",
			test: func(t *testing.T) {
				jwtToken := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					jwt.MapClaims{
						"sub": "login",
					})

				user := models.User{ID: "user-id", Login: "user", Hash: "hash"}

				authServiceMock.EXPECT().GetUser(gomock.Any(), "login").Return(&user, nil)
				jwtServiceMock.EXPECT().ValidateToken("token").Return(jwtToken, nil)
				balanceServiceMock.EXPECT().GetWithdrawalFlow(gomock.Any(), "user-id").Return([]models.WithdrawalFlowItem{
					{
						OrderID:     "order-id",
						Sum:         123.123,
						ProcessedAt: utils.RFC3339Date{Time: time.Date(2009, 11, 17, 0, 0, 0, 0, time.UTC)},
					},
				}, nil)
			},
			expectedCode:    http.StatusOK,
			expectedMessage: "[{\"order\":\"order-id\",\"sum\":123.123,\"processed_at\":\"2009-11-17T00:00:00Z\"}]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.test != nil {
				tc.test(t)
			}

			res, mes := utils.TestRequest(
				t,
				testServer,
				tc.methodName,
				tc.targetURL,
				map[string]string{"Authorization": "Bearer token"},
				nil,
			)
			res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedMessage, mes)
		})
	}
}
