package dsutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DataWorkbench/gproto/xgo/types/pbmodel"
	"github.com/DataWorkbench/gproto/xgo/types/pbmodel/pbdatasource"
	"github.com/Shopify/sarama"
	"github.com/dazheng/gohive"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/dutchcoders/goftp"
	"github.com/go-redis/redis"
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
	elastic6 "github.com/olivere/elastic/v6"
	elastic7 "github.com/olivere/elastic/v7"
	"github.com/pkg/sftp"
	"github.com/samuel/go-zookeeper/zk"
	_ "github.com/sijms/go-ora/v2"
	"golang.org/x/crypto/ssh"
	"gopkg.in/mgo.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func pingNetwork(address string, port int32) (err error) {
	var conn net.Conn
	ip := net.JoinHostPort(address, strconv.Itoa(int(port)))
	conn, err = net.DialTimeout("tcp", ip, time.Second*5)
	if err != nil {
		return err
	}
	if conn != nil {
		_ = conn.Close()
	}
	return
}

func pingMysql(url *pbdatasource.MySQLURL) (err error) {
	if err = pingNetwork(url.Host, url.Port); err != nil {
		return
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		url.User, url.Password, url.Host, url.Port, url.Database,
	)
	var db *gorm.DB
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	} else {
		if sqlDB, e := db.DB(); e == nil {
			_ = sqlDB.Close()
		}
	}
	return nil
}

func pingOracle(url *pbdatasource.OracleURL) (err error) {
	if err = pingNetwork(url.Host, url.Port); err != nil {
		return
	}

	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d?SID=%s", url.User, url.Password, url.Host, url.Port, url.Database)

	var db *sql.DB
	db, err = sql.Open("oracle", connStr)

	if err != nil {
		return err
	} else {
		if e := db.Ping(); e == nil {
			_ = db.Close()
		} else {
			return e
		}
	}
	return nil
}

func pingOceanBaseURL(url *pbdatasource.OceanBaseURL) (err error) {
	if err = pingNetwork(url.Host, url.Port); err != nil {
		return
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		url.User, url.Password, url.Host, url.Port, url.Database,
	)
	var db *gorm.DB
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	} else {
		if sqlDB, e := db.DB(); e == nil {
			_ = sqlDB.Close()
		}
	}
	return nil
}
func pingWebApi(url *pbdatasource.WebApiURL) (err error) {
	weburl := fmt.Sprintf(
		"http://%s:%d/%s",
		url.Host, url.Port, url.Url) // 替换为你的API URL
	client := http.Client{
		Timeout: 30 * time.Second, // 设置超时时间
	}

	resp, err := client.Head(weburl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//if resp.StatusCode == http.StatusOK {
	return nil
	//}
}

func pingMqtt(url *pbdatasource.MqttURL) (err error) {
	// 替换为你的MQTT服务器地址和端口
	host := fmt.Sprintf(
		"%s:%d",
		url.Host, url.Port)
	timeout := 30 * time.Second

	// 尝试建立TCP连接
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func pingPostgreSQL(url *pbdatasource.PostgreSQLURL) (err error) {
	if err = pingNetwork(url.Host, url.Port); err != nil {
		return
	}

	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d  dbname=%s ",
		url.User, url.Password, url.Host, url.Port, url.Database,
	)
	var db *gorm.DB
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	} else {
		if sqlDB, e := db.DB(); e == nil {
			_ = sqlDB.Close()
		}
	}
	return nil
}

func pingSqlServer(url *pbdatasource.SqlServerURL) (err error) {
	if err = pingNetwork(url.Host, url.Port); err != nil {
		return
	}

	connString := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d;encrypt=disable", url.Host, url.Database, url.User, url.Password, url.Port)
	db, err := sql.Open("mssql", connString)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func pingClickHouse(url *pbdatasource.ClickHouseURL) (err error) {
	if err = pingNetwork(url.Host, url.Port); err != nil {
		return
	}

	connect, err := dbr.Open("clickhouse", fmt.Sprintf("http://%s:%s@%s:%d/%s", url.User, url.Password, url.Host, url.Port, url.Database), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = connect.Close()
	}()
	err = connect.Ping()
	if err != nil {
		return err
	}
	return
}

func pingKafka(url *pbdatasource.KafkaURL) (err error) {
	var brokes []string
	for _, value := range url.KafkaBrokers {
		brokes = append(brokes, fmt.Sprintf("%s:%d", value.Host, value.Port))
	}
	config := sarama.NewConfig()
	config.Version = sarama.V0_10_1_1
	client, err := sarama.NewClient(brokes, config)
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}

func pingHBase(url *pbdatasource.HBaseURL) (err error) {
	var conn *zk.Conn

	config := make(map[string]string)
	if err = json.Unmarshal([]byte(url.Config), &config); err != nil {
		return err
	}

	zkHosts := config["hbase.zookeeper.quorum"]
	zkPort := config["hbase.zookeeper.property.clientPort"]
	if zkPort == "" {
		zkPort = "2181"
	}

	servers := make([]string, 0, len(zkHosts))
	for _, node := range strings.Split(zkHosts, ",") {
		servers = append(servers, fmt.Sprintf("%s:%s", node, zkPort))
	}

	conn, _, err = zk.Connect(servers, time.Millisecond*100)
	if err != nil {
		return err
	}
	_, _, _, err = conn.ExistsW("/")
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func pingFtp(url *pbdatasource.FtpURL) (err error) {
	if url.Protocol.Number() == 1 {
		//ftp
		conn, err := goftp.Connect(fmt.Sprintf("%v:%d", url.Host, url.Port))
		if err != nil {
			return err
		}
		err = conn.Login(url.User, url.Password)
		if err != nil {
			return err
		}
		_ = conn.Close()
		return nil
	}
	//sftp
	/*homePath, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	key, err := ioutil.ReadFile(path.Join(homePath, ".ssh", "id_rsa"))
	if err != nil {
		fmt.Println(err)
	}*/
	signer, err := ssh.ParsePrivateKey([]byte(url.PrivateKey))
	if err != nil {
		return err
	}
	config := &ssh.ClientConfig{
		User: url.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(url.Password),
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", url.Host, config)
	if err != nil {
		return err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	// Close connection
	defer client.Close()
	defer conn.Close()
	return nil
}

func pingHDFS(url *pbdatasource.HDFSURL) (err error) {
	var conn net.Conn
	// https://github.com/colinmarc/hdfs -- install the hadoop client. so don't use it.
	// https://studygolang.com/articles/766 -- use 50070 http port. but user input the IPC port.
	ip := net.JoinHostPort(url.NameNode, strconv.Itoa(int(url.Port)))
	conn, err = net.DialTimeout("tcp", ip, time.Millisecond*2000)
	if err != nil {
		return err
	}
	if conn != nil {
		_ = conn.Close()
	}
	return nil
}

func pingHive(url *pbdatasource.HiveURL) (err error) {
	conn, err := gohive.Connect(fmt.Sprintf("%s:%d", url.Host, url.Port), gohive.DefaultOptions)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func pingElasticSearch(url *pbdatasource.ElasticSearchURL) (err error) {
	if url.Version[0:1] == "6" {
		_, err := elastic6.NewClient(
			elastic6.SetSniff(false),
			elastic6.SetURL(fmt.Sprintf("http://%s:%d/", url.Host, url.Port)),
			elastic6.SetBasicAuth(url.User, url.Password),
		)

		if err != nil {
			return err
		}
	}
	if url.Version[0:1] == "7" {
		_, err := elastic7.NewClient(
			elastic7.SetSniff(false),
			elastic7.SetURL(fmt.Sprintf("http://%s:%d/", url.Host, url.Port)),
			elastic7.SetBasicAuth(url.User, url.Password),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func pingMongoDb(url *pbdatasource.MongoDbURL) (err error) {
	session, err := mgo.Dial(fmt.Sprintf("%s:%d", url.Hosts[0].Host, url.Hosts[0].Port))
	if err != nil {
		return err
	}
	session.SetMode(mgo.Monotonic, true)
	if url.User != "" && url.Password != "" {
		db := session.DB("admin")
		err = db.Login(url.User, url.Password)
		if err != nil {
			return err
		}
	}
	defer session.Close()
	return nil
}

func pingRedis(url *pbdatasource.RedisURL) (err error) {
	var redisOption = redis.Options{
		Addr: fmt.Sprintf("%s:%d", url.Hosts[0].Host, url.Hosts[0].Port),
	}
	if url.Password != "" {
		redisOption.Password = url.Password
	}

	client := redis.NewClient(&redisOption)
	_, err = client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func PingDataSourceConnection(ctx context.Context, sourceType pbmodel.DataSource_Type, sourceURL *pbmodel.DataSource_URL) (connInfo *pbmodel.DataSourceConnection, err error) {
	begin := time.Now()
	message := ""
	result := pbmodel.DataSourceConnection_Success

	switch sourceType {
	case pbmodel.DataSource_MySQL:
		err = pingMysql(sourceURL.Mysql)
	case pbmodel.DataSource_PostgreSQL:
		err = pingPostgreSQL(sourceURL.Postgresql)
	case pbmodel.DataSource_SqlServer:
		err = pingSqlServer(sourceURL.Sqlserver)
	case pbmodel.DataSource_Oracle:
		err = pingOracle(sourceURL.Oracle)
	case pbmodel.DataSource_DB2:
		//empty
	case pbmodel.DataSource_SapHana:
		//empty
	case pbmodel.DataSource_Kafka:
		err = pingKafka(sourceURL.Kafka)
	case pbmodel.DataSource_S3:
		//empty
	case pbmodel.DataSource_ClickHouse:
		err = pingClickHouse(sourceURL.Clickhouse)
	case pbmodel.DataSource_HBase:
		err = pingHBase(sourceURL.Hbase)
	case pbmodel.DataSource_Ftp:
		err = pingFtp(sourceURL.Ftp)
	case pbmodel.DataSource_HDFS:
		err = pingHDFS(sourceURL.Hdfs)
	case pbmodel.DataSource_Hive:
		err = pingHive(sourceURL.Hive)
	case pbmodel.DataSource_ElasticSearch:
		err = pingElasticSearch(sourceURL.ElasticSearch)
	case pbmodel.DataSource_MongoDb:
		err = pingMongoDb(sourceURL.MongoDb)
	case pbmodel.DataSource_Redis:
		err = pingRedis(sourceURL.Redis)
	case pbmodel.DataSource_OceanBase:
		err = pingOceanBaseURL(sourceURL.Oceanbase)
	case pbmodel.DataSource_Mqtt:
		err = pingMqtt(sourceURL.Mqtt)
	case pbmodel.DataSource_WebApi:
		err = pingWebApi(sourceURL.Webapi)
	}

	if err != nil {
		result = pbmodel.DataSourceConnection_Failed
		message = err.Error()
		err = nil
	}

	connInfo = &pbmodel.DataSourceConnection{
		SpaceId:     "",
		SourceId:    "",
		NetworkId:   "",
		Status:      pbmodel.DataSourceConnection_Enabled,
		Result:      result,
		Message:     message,
		Created:     begin.Unix(),
		Elapse:      time.Since(begin).Milliseconds(),
		NetworkInfo: nil,
	}
	return
}
