package Cassandra

import (
	"github.com/gocql/gocql"
	"time"
	"math/rand"
	"log"
	"sync"
)

// Session holds our connection to Cassandra
var Session *gocql.Session

func init() {
	time.Sleep(30 * time.Second)
	var err error

	cluster := gocql.NewCluster("cassandra")
	
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	if err := Session.Query(`
		CREATE KEYSPACE cortex WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};`).Exec(); err != nil {
			log.Print(err)
	} 
	cluster.Keyspace = "cortex"
	cluster.ProtoVersion = 3
	cluster.Consistency = gocql.One
	
	log.Print("Cassandra init done")

}

func CreateSchema() {
	if err := Session.Query(`
		create table cortex.chunk(hash text, range blob, value blob, PRIMARY KEY(hash, range));`).Exec(); err != nil {
			log.Print(err)
	} 
	log.Print("Schema Created")
}

func InsertData(doneCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select{
		case <-doneCh:
			log.Print("Stopping data inserting")
			return
		default:
			if err := Session.Query(`
      			INSERT INTO cortex.chunk (hash, range, value) VALUES (?, ?, ?)`,
	  			randomString(100), randomString(4), randomString(1000)).Exec(); err != nil {
				log.Fatal(err)
			}
			log.Print("Data created!")
		}
	}
}

func ReadData(doneCh chan struct{}, wg *sync.WaitGroup) {
	log.Print("Reading data")
	defer wg.Done()
	for {
		select{
		case <-doneCh:
			log.Print("Stopping data reading")
			return
		default:
			var counter int
			var hash string
			if err := Session.Query(`SELECT COUNT(*), hash FROM cortex.chunk`).Scan(&counter, &hash); err != nil {
				log.Fatal(err)
			}
			log.Printf("Readed %d rows!", counter)
		}
	}
}

func randomInt(min, max int) int {
    return min + rand.Intn(max-min)
}

func randomString(len int) string {
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        bytes[i] = byte(randomInt(33, 122))
    }
    return string(bytes)
}