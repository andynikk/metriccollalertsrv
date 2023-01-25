package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andynikk/metriccollalertsrv/internal/compression"
	"github.com/andynikk/metriccollalertsrv/internal/cryptohash"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
	"github.com/andynikk/metriccollalertsrv/internal/encryption"
	"github.com/andynikk/metriccollalertsrv/internal/environment"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"github.com/andynikk/metriccollalertsrv/internal/repository"
	"github.com/go-chi/chi/v5"
)

func TestFuncServerHTTP(t *testing.T) {
	var fValue float64 = 0.001
	var iDelta int64 = 10

	config := environment.InitConfigServer()
	server, err := handlers.NewServer(config)
	if err != nil {
		fmt.Println("not create server")
		return
	}
	repStore := server.GetRepStore()
	router := server.(*handlers.ServerHTTP).Router

	t.Run("Checking init server", func(t *testing.T) {
		//rp.Config = environment.InitConfigServer()
		if repStore.Config.Address == "" {
			t.Errorf("Error checking init server")
		}
	})
	fmt.Println(repStore.Config.Address)

	repStore.MutexRepo = make(repository.MutexRepo)
	t.Run("Checking init router", func(t *testing.T) {
		if router == nil {
			t.Errorf("Error checking init router")
		}
	})

	t.Run("Checking init config", func(t *testing.T) {
		if repStore.Config.Address == "" {
			t.Errorf("Error checking init config")
		}
	})

	postStr := fmt.Sprintf("http://%s/update/gauge/Alloc/0.1\nhttp://%s/update/gauge/BuckHashSys/0.002"+
		"\nhttp://%s/update/counter/PollCount/5", repStore.Config.Address, repStore.Config.Address, repStore.Config.Address)

	t.Run("Checking the filling of metrics Gauge", func(t *testing.T) {

		messageRaz := strings.Split(postStr, "\n")
		if len(messageRaz) != 3 {
			t.Errorf("The string (%s) was incorrectly decomposed into an array", postStr)
		}
	})

	t.Run("Checking rsa crypt", func(t *testing.T) {
		t.Run("Checking init crypto key", func(t *testing.T) {
			repStore.PK, _ = encryption.InitPrivateKey(repStore.Config.CryptoKey)
			if repStore.Config.CryptoKey != "" && repStore.PK == nil {
				t.Errorf("Error checking init crypto key")
			}
		})
		t.Run("Checking rsa encrypt", func(t *testing.T) {
			testMsg := "Тестовое сообщение"

			encryptMsg, err := repStore.PK.RsaEncrypt([]byte(testMsg))
			if err != nil {
				t.Errorf("Error checking rsa encrypt")
			}
			t.Run("Checking rsa decrypt", func(t *testing.T) {
				decryptMsg, err := repStore.PK.RsaDecrypt(encryptMsg)
				if err != nil {
					t.Errorf("Error checking rsa decrypt")
				}
				byteTestMsg := []byte(testMsg)
				if !bytes.Equal(decryptMsg, byteTestMsg) {
					t.Errorf("Error checking rsa decrypt")
				}
			})
		})
	})

	t.Run("Checking connect DB", func(t *testing.T) {
		t.Run("Checking create DB table", func(t *testing.T) {
			if config.DatabaseDsn != "" {
				repStore.Config.Storage = repository.NewStorage(config.DatabaseDsn, config.StoreFile)
				t.Run("Checking handlers /ping GET", func(t *testing.T) {
					err = config.Storage.ConnDB()
					if err != nil {
						t.Errorf("Error handlers 1 /ping GET")
					}
				})
			}
		})
	})

	t.Run("Checking metric methods", func(t *testing.T) {
		t.Run(`Checking method "String" type "Counter"`, func(t *testing.T) {
			var c repository.Counter = 58
			if c.String() != "58" {
				t.Errorf(`Error method "String" for Counter `)
			}
		})
		t.Run(`Checking method "String" type "Gauge"`, func(t *testing.T) {
			var c repository.Gauge = 0.001
			if c.String() != "0.001" {
				t.Errorf(`Error method "String" for Counter `)
			}
		})
		t.Run(`Checking method "GetMetrics" type "Counter"`, func(t *testing.T) {
			mType := "counter"
			id := "TestCounter"
			hashKey := "TestHash"

			c := repository.Counter(58)

			mt := c.GetMetrics(mType, id, hashKey)
			msg := fmt.Sprintf("MType: %s, ID: %s, Value: %v, Delta: %d, Hash: %s",
				mt.MType, mt.ID, 0, *mt.Delta, mt.Hash)

			if msg != "MType: counter, ID: TestCounter, Value: 0, Delta: 58, Hash: 29bd8e4bde7ec6302393fe3f7954895a65f4d4b22372d00a35fc1adbcc2ec239" {
				t.Errorf(`method "GetMetrics" type "Counter"`)
			}
		})
		t.Run(`Checking method "GetMetrics" type "Gauge"`, func(t *testing.T) {
			mType := "gauge"
			id := "TestGauge"
			hashKey := "TestHash"

			g := repository.Gauge(0.01)

			mt := g.GetMetrics(mType, id, hashKey)
			msg := fmt.Sprintf("MType: %s, ID: %s, Value: %f, Delta: %d, Hash: %s",
				mt.MType, mt.ID, *mt.Value, 0, mt.Hash)
			if msg != "MType: gauge, ID: TestGauge, Value: 0.010000, Delta: 0, Hash: 4e5d8a0e257dd12355b15f730591dddd9e45e18a6ef67460a58f20edc12c9465" {
				t.Errorf(`method "GetMetrics" type "Counter"`)
			}
		})
		t.Run(`Checking method "Set" type "Counter"`, func(t *testing.T) {
			var c repository.Counter
			var i int64 = 58
			v := encoding.Metrics{
				ID:    "",
				MType: "",
				Delta: &i,
				Hash:  "",
			}
			c.Set(v)

			if c != 58 {
				t.Errorf(`Error method "Set" for Counter `)
			}
		})
		t.Run(`Checking method "Set" type "Gauge"`, func(t *testing.T) {
			var g repository.Gauge
			var f float64 = 0.01

			v := encoding.Metrics{
				ID:    "",
				MType: "",
				Value: &f,
				Hash:  "",
			}
			g.Set(v)
			if g != 0.01 {
				t.Errorf(`Error method "Set" for Counter `)
			}
		})
		t.Run(`Checking method "Type" type "Counter"`, func(t *testing.T) {
			var c repository.Counter = 58
			if c.Type() != "counter" {
				t.Errorf(`Error method "Type" for Counter `)
			}
		})
		t.Run(`Checking method "Type" type "Gauge"`, func(t *testing.T) {
			var g repository.Gauge = 0.001
			if g.Type() != "gauge" {
				t.Errorf(`Error method "Type" for Counter `)
			}
		})
		t.Run(`Checking method "SetFromText" type "Counter"`, func(t *testing.T) {
			metValue := "58"
			c := repository.Gauge(0)
			c.SetFromText(metValue)
			if c != 58 {
				t.Errorf(`Error method "SetFromText" for Counter `)
			}
		})
		t.Run(`Checking method "SetFromText" type "Gauge"`, func(t *testing.T) {
			metValue := "0.01"
			g := repository.Gauge(0)
			g.SetFromText(metValue)
			if g != 0.01 {
				t.Errorf(`Error method "SetFromText" for Counter `)
			}
		})

		t.Run(`Checking method PrepareDataForBackup`, func(t *testing.T) {
			valG := repository.Gauge(0)
			if ok := valG.SetFromText("0.001"); !ok {
				t.Errorf(`Error method "PrepareDataForBackup"`)
			}
			repStore.MutexRepo["TestGauge"] = &valG

			valC := repository.Counter(0)
			if ok := valC.SetFromText("58"); !ok {
				t.Errorf(`Error method "PrepareDataForBackup"`)
			}
			repStore.MutexRepo["TestCounter"] = &valC

			data := repStore.PrepareDataBuckUp()
			if len(data) != 2 {
				t.Errorf(`Error method "PrepareDataForBackup"`)
			}
		})
	})

	t.Run("Checking handlers", func(t *testing.T) {
		ts := httptest.NewServer(router)
		defer ts.Close()

		t.Run("Checking handler /update/{metType}/{metName}/{metValue}/", func(t *testing.T) {
			resp := testRequest(t, ts, http.MethodPost, "/update/gauge/TestGauge/0.01", nil)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Error handler /update/{metType}/{metName}/{metValue}/")
			}
			t.Run("Checking handler /value/", func(t *testing.T) {
				resp := testRequest(t, ts, http.MethodGet, "/value/gauge/TestGauge", nil)
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Error handler /value/")
				}
			})
		})
		//t.Run("Checking handler /updates POST/", func(t *testing.T) {
		//	//resp := testRequest(t, ts, http.MethodPost, "/updates", nil)
		//	//defer resp.Body.Close()
		//	//
		//	//if resp.StatusCode != http.StatusOK {
		//	//	t.Errorf("Error handler //update/{metType}/{metName}/{metValue}/")
		//	//}
		//})
		t.Run("Checking handler /update POST/", func(t *testing.T) {
			//testA := testArray("")
			testA := testMericGougeHTTP("")
			arrMetrics, err := json.MarshalIndent(testA, "", " ")
			if err != nil {
				t.Errorf("Error handler /update POST/")
			}
			body := bytes.NewReader(arrMetrics)
			resp := testRequest(t, ts, http.MethodPost, "/update", body)
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Error handler /update POST/")
			}
			t.Run("Checking handler /value POST/", func(t *testing.T) {
				metricJSON, err := json.MarshalIndent(testMericGougeHTTP(""), "", " ")
				if err != nil {
					t.Errorf("Error handler /value POST/")
				}
				body := bytes.NewReader(metricJSON)

				resp := testRequest(t, ts, http.MethodPost, "/value", body)
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Error handler /value POST/")
				}
			})
		})
	})

	t.Run("Checking the filling of metrics", func(t *testing.T) {
		t.Run("Checking the type of the first line", func(t *testing.T) {
			var typeGauge = "gauge"

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[0]

			if strings.Contains(valElArr, typeGauge) == false {
				t.Errorf("The Gauge type was incorrectly determined")
			}
		})

		tests := []struct {
			name           string
			request        string
			wantStatusCode int
		}{
			{name: "Проверка на установку значения counter", request: "/update/counter/testSetGet332/6",
				wantStatusCode: http.StatusOK},
			{name: "Проверка на не правильный тип метрики", request: "/update/notcounter/testSetGet332/6",
				wantStatusCode: http.StatusNotImplemented},
			{name: "Проверка на не правильное значение метрики", request: "/update/counter/testSetGet332/non",
				wantStatusCode: http.StatusBadRequest},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				r := chi.NewRouter()
				ts := httptest.NewServer(r)

				smm := new(repository.SyncMapMetrics)
				smm.MutexRepo = make(repository.MutexRepo)

				r.Post("/update/{metType}/{metName}/{metValue}", repStore.HandlerSetMetricaPOST)

				defer ts.Close()
				resp := testRequest(t, ts, http.MethodPost, tt.request, nil)
				defer resp.Body.Close()

				if resp.StatusCode != tt.wantStatusCode {
					t.Errorf("Ответ не верен")
				}
			})
		}
	})

	t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
		t.Run("Checking the filling of metrics Counter", func(t *testing.T) {
			var typeCounter = "counter"

			messageRaz := strings.Split(postStr, "\n")
			valElArr := messageRaz[2]

			if strings.Contains(valElArr, typeCounter) == false {
				t.Errorf("The Counter type was incorrectly determined")
			}
		})

	})

	t.Run("Checking compresion - decompression", func(t *testing.T) {

		textGzip := "Testing massage"
		arrByte := []byte(textGzip)

		compresMsg, err := compression.Compress(arrByte)
		if err != nil {
			t.Errorf("Error compres")
		}

		decompresMsg, err := compression.Decompress(compresMsg)
		if err != nil {
			t.Errorf("Error decompres")
		}

		msgReader := bytes.NewReader(decompresMsg)
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(msgReader); err != nil {
			t.Errorf("Error read decompression msg")
		}
		decompresText := buf.String()

		if decompresText != textGzip {
			t.Errorf("Error checking compresion - decompression")
		}
	})

	t.Run("Checking Hash SHA 256", func(t *testing.T) {
		configKey := "testKey"
		txtData := "Test data"

		hashData := cryptohash.HashSHA256(txtData, configKey)
		if hashData == "" || len(hashData) != 64 {
			t.Errorf("Error checking Hash SHA 256")
		}

		t.Run("Checking set val in map", func(t *testing.T) {
			rs := new(handlers.RepStore)

			//smm := new(repository.SyncMapMetrics)
			//smm.MutexRepo = make(repository.MutexRepo)

			rs.MapMetrics.MutexRepo = make(repository.MutexRepo)

			arrM := testArray(configKey)

			for idx, val := range arrM {
				if idx == 0 {
					valG := repository.Gauge(0)
					rs.MutexRepo[val.ID] = &valG
				} else {
					valC := repository.Counter(0)
					rs.MutexRepo[val.ID] = &valC
				}
				rs.MutexRepo[val.ID].Set(val)
			}

			erorr := false
			for idx, val := range rs.MutexRepo {
				gauge := repository.Gauge(fValue)
				counter := repository.Counter(iDelta)
				if idx == "TestGauge" && val.String() != gauge.String() {
					erorr = true
				} else if idx == "TestCounter" && val.String() != counter.String() {
					erorr = true
				}
			}

			if erorr {
				t.Errorf("Error checking Hash SHA 256")
			}
		})
	})

	t.Run("Checking marshal metrics JSON", func(t *testing.T) {

		for key, val := range repStore.MutexRepo {
			mt := val.GetMetrics(val.Type(), key, repStore.Config.Key)
			_, err := mt.MarshalMetrica()
			if err != nil {
				t.Errorf("Error checking marshal metrics JSON")
			}
		}
	})
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	defer resp.Body.Close()

	return resp
}

func testArray(configKey string) encoding.ArrMetrics {
	var arrM encoding.ArrMetrics

	arrM = append(arrM, testMericGougeHTTP(configKey))
	arrM = append(arrM, testMericCaunterHTTP(configKey))

	return arrM
}

func testMericGougeHTTP(configKey string) encoding.Metrics {

	var fValue float64 = 0.001

	var mGauge encoding.Metrics
	mGauge.ID = "TestGauge"
	mGauge.MType = "gauge"
	mGauge.Value = &fValue
	mGauge.Hash = cryptohash.HashSHA256(mGauge.ID, configKey)

	return mGauge
}

func testMericCaunterHTTP(configKey string) encoding.Metrics {
	var iDelta int64 = 10

	var mCounter encoding.Metrics
	mCounter.ID = "TestCounter"
	mCounter.MType = "counter"
	mCounter.Delta = &iDelta
	mCounter.Hash = cryptohash.HashSHA256(mCounter.ID, configKey)

	return mCounter
}
