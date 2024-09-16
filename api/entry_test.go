package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "bitbucket.org/jessyw/go_simplebank/db/mock"
	db "bitbucket.org/jessyw/go_simplebank/db/sqlc"
	"bitbucket.org/jessyw/go_simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestFindEntryByAccountIDAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	entry := randomEntry(account)

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Entry{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				// build stubs
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateEntryAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)
	entry := randomEntry(account)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"account_id": account.ID,
				"amount":     entry.Amount,
				"created_at": entry.CreatedAt,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateEntryParams{
					AccountID: account.ID,
					Amount:    entry.Amount,
				}

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					CreateEntry(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(entry, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name: "AccountNotFound",
			body: gin.H{
				"account_id": account.ID,
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().CreateEntry(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalErrorOnGetAccount",
			body: gin.H{
				"account_id": account.ID,
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().CreateEntry(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InternalErrorOnCreateEntry",
			body: gin.H{
				"account_id": account.ID,
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					CreateEntry(gomock.Any(), gomock.Eq(db.CreateEntryParams{
						AccountID: account.ID,
						Amount:    entry.Amount,
					})).
					Times(1).
					Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidRequest",
			body: gin.H{
				"account_id": 0, // Invalid account ID
				"amount":     entry.Amount,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().CreateEntry(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/entries"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListEntriesAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	account := randomAccount(user.Username)
	entries := make([]db.Entry, n)
	for i := 0; i < n; i++ {
		entries[i] = randomEntry(account)
	}

	type Query struct {
		id     int64
		offset int
		limit  int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				id:     account.ID,
				offset: 1,
				limit:  n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: account.ID,
					Limit:     int32(n),
					Offset:    0,
				}

				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(entries, nil)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
				requireBodyMatchEntries(t, recoder.Body, entries)
			},
		},
		{
			name: "NotFound",
			query: Query{
				id:     account.ID,
				offset: 1,
				limit:  n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEntriesParams{
					AccountID: account.ID,
					Limit:     int32(n),
					Offset:    0,
				}

				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Entry{}, sql.ErrNoRows) // Return empty list
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recoder.Code)
			},
		},
		{
			name: "InvalidOffset",
			query: Query{
				id:     account.ID,
				offset: 0, // Invalid offset
				limit:  n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Any()).
					Times(0) // Shouldn't reach the store
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
		{
			name: "InvalidLimit",
			query: Query{
				id:     account.ID,
				offset: 1,
				limit:  15, // Invalid limit
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Any()).
					Times(0) // Shouldn't reach the store
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/entries"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			q := request.URL.Query()
			q.Add("id", fmt.Sprintf("%d", tc.query.id))
			q.Add("offset", fmt.Sprintf("%d", tc.query.offset))
			q.Add("limit", fmt.Sprintf("%d", tc.query.limit))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}

}

func randomEntry(account db.Account) db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: account.ID,
		Amount:    util.RandomMoney(),
		CreatedAt: time.Now(),
	}
}

func requireBodyMatchEntry(t *testing.T, body *bytes.Buffer, entry db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntry db.Entry
	err = json.Unmarshal(data, &gotEntry)
	require.NoError(t, err)
	require.Equal(t, entry.ID, gotEntry.ID)
	require.Equal(t, entry.AccountID, gotEntry.AccountID)
	require.Equal(t, entry.Amount, gotEntry.Amount)

	// compare  date without check  internal fields of time.Time
	require.WithinDuration(t, entry.CreatedAt, gotEntry.CreatedAt, time.Second)
}

func requireBodyMatchEntries(t *testing.T, body *bytes.Buffer, entries []db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntries []db.Entry
	err = json.Unmarshal(data, &gotEntries)
	require.NoError(t, err)
	require.Equal(t, len(entries), len(gotEntries))

	for i := range entries {
		require.Equal(t, entries[i].ID, gotEntries[i].ID)
		require.Equal(t, entries[i].AccountID, gotEntries[i].AccountID)
		require.Equal(t, entries[i].Amount, gotEntries[i].Amount)

		require.WithinDuration(t, entries[i].CreatedAt, gotEntries[i].CreatedAt, time.Second)
	}
}
