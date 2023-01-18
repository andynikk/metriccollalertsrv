package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
	"github.com/andynikk/metriccollalertsrv/internal/encoding"
)

type Context struct {
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

// DBConnector структура хранения конекта с базой данной
type DBConnector struct {
	Pool    *pgxpool.Pool
	Context Context
}

type transitMetrics struct {
	MType string
	ID    string
	Value *float64
	Delta *int64
	Hash  string
}

type arrTransitMetrics struct {
	Arr []transitMetrics
}

// PoolDB создает коннект с базой данных.
// Хранит коннект (pgxpool.Connect) в настройках.
// Из конекта создает Pool, при необходимости.
func PoolDB(dsn string) (*DBConnector, error) {
	if dsn == "" {
		return new(DBConnector), errors.New("пустой путь к базе")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		fmt.Println("++++++++++++++014", err)
		fmt.Print(err.Error())
	}

	strQuery := fmt.Sprintf(constants.QueryCheckExistDB, constants.NameDB)
	rows, err := pool.Query(ctx, strQuery)
	if err != nil {
		fmt.Println("++++++++++++++015", err)
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		strQuery = fmt.Sprintf(constants.QueryDB, constants.NameDB)
		if _, err = pool.Exec(ctx, strQuery); err != nil {
			fmt.Println("++++++++++++++016", err)
			return nil, err

		}
	}

	fmt.Println("++++++++++++++017-1", dsn)
	dsn = strings.Replace(dsn, "/"+constants.NameDB, "", -1)
	fmt.Println("++++++++++++++017-2", dsn)
	fmt.Println("++++++++++++++017-3", dsn+"/"+constants.NameDB)
	//pool, err = pgxpool.Connect(ctx, dsn+"/"+constants.NameDB)
	pool, err = pgxpool.Connect(ctx, dsn)
	if err != nil {
		fmt.Println("++++++++++++++017", err)
		return nil, err
	}

	dbc := DBConnector{
		Pool: pool,
		Context: Context{
			Ctx:        ctx,
			CancelFunc: cancelFunc,
		},
	}
	fmt.Println("++++++++++++++018", pool)
	return &dbc, nil
}

// SetMetric2DB Добавляет метрики в БД.
// Из массива (тип encoding.ArrMetrics) создает запрос к БД по имени и типу метрики.
// По найденным метрикам создает набор SQL-запросов update
// По не найденным метрикам создает набор SQL-запросов insert
// Далает вызов БД один раз, сразу по всем update &  insert
func (DataBase *DBConnector) SetMetric2DB(storedData encoding.ArrMetrics) error {

	ctx := context.Background()
	conn, err := DataBase.Pool.Acquire(ctx)

	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		constants.Logger.ErrorLog(err)
	}

	allWhereVal := ""

	var allTM []transitMetrics
	for _, data := range storedData {
		if allWhereVal != "" {
			allWhereVal = allWhereVal + " or "
		}
		allWhereVal = allWhereVal + fmt.Sprintf(
			`("MType" = '%s' and "ID" = '%s')`,
			data.MType, data.ID)

		allTM = append(allTM, transitMetrics{
			MType: data.MType,
			ID:    data.ID,
			Value: data.Value,
			Delta: data.Delta,
			Hash:  data.Hash,
		})
	}
	allArrTM := new(arrTransitMetrics)
	allArrTM.Arr = allTM

	if allWhereVal == "" {
		return nil
	}
	allWhereVal = "(" + allWhereVal + ")"
	txtQuery := fmt.Sprintf(`SELECT * FROM metrics.store WHERE %s;`, allWhereVal)

	var updTM []transitMetrics
	rows, err := conn.Query(ctx, txtQuery)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return err
	}
	for rows.Next() {
		var d encoding.Metrics

		err = rows.Scan(&d.ID, &d.MType, &d.Value, &d.Delta, &d.Hash)
		if err != nil {
			constants.Logger.ErrorLog(err)
			continue
		}

		updTM = append(updTM, transitMetrics{MType: d.MType, ID: d.ID})
	}
	updArrTM := new(arrTransitMetrics)
	updArrTM.Arr = updTM

	txtQueryUpdata := ""
	txtQueryInsert := ""

	for _, val := range allArrTM.Arr {
		sValue := fmt.Sprintf("%d", 0)
		if val.Value != nil {
			sValue = fmt.Sprintf("%f", *val.Value)
		}
		sDelta := fmt.Sprintf("%d", 0)
		if val.Delta != nil {
			sDelta = fmt.Sprintf("%d", *val.Delta)
		}

		if ok := updArrTM.find(val.MType, val.ID); ok {
			if txtQueryUpdata != "" {
				txtQueryUpdata = txtQueryUpdata + "\n"
			}
			txtQueryUpdata = txtQueryUpdata + fmt.Sprintf(
				`UPDATE metrics.store SET "Value"=%s, "Delta"=%s, "Hash"='%s' WHERE	"ID" = '%s'	and "MType" = '%s';`,
				sValue, sDelta, val.Hash, val.ID, val.MType)
			continue
		}

		if txtQueryInsert != "" {
			txtQueryInsert = txtQueryInsert + "\n"
		}
		txtQueryInsert = txtQueryInsert + fmt.Sprintf(
			`INSERT INTO metrics.store ("ID", "MType", "Value", "Delta", "Hash") VALUES ('%s', '%s', %v, %s, '%s');`,
			val.ID, val.MType, sValue, sDelta, val.Hash)
	}

	txtExec := txtQueryInsert + "\n" + txtQueryUpdata
	if _, err := conn.Exec(ctx, txtExec); err != nil {
		conn.Release()
		return errors.New("ошибка изменения данных в БД")
	}

	if err := tx.Commit(ctx); err != nil {
		constants.Logger.ErrorLog(err)
	}

	conn.Release()
	ctx.Done()

	return nil
}

func (atm arrTransitMetrics) find(mtype string, id string) bool {
	for _, val := range atm.Arr {
		if val.MType == mtype && val.ID == id {
			return true
		}
	}
	return false
}
