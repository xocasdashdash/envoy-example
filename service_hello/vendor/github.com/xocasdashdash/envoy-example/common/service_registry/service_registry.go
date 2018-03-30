package service_registry

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

type ServiceRegistry struct {
	Location string
	Retries  int
}

func (sr ServiceRegistry) Register(serviceName, ip string, port int) error {
	config := api.DefaultConfig()
	config.Address = sr.Location
	retriesLeft := sr.Retries
	var srvRegistrator *api.Agent
	for {
		log.Printf("Trying to connect to consul...")
		client, _ := api.NewClient(config)
		srvRegistrator = client.Agent()
		//Check connection to consul
		_, err := srvRegistrator.Self()

		if err != nil && retriesLeft <= 0 {
			log.Printf("Encountered error connecting to consul on %s => %s\n", "consul", err)
			return err
		} else if err != nil {
			waitTime := time.Duration(math.Exp2(float64(sr.Retries-retriesLeft))) * time.Second
			log.Printf("Encountered error connecting to consul on %s => %s\n", "consul", err)
			log.Printf("Waiting %.0f secs until next connection attempt", waitTime.Seconds())
			log.Printf("Retries left: %d", retriesLeft)
			retriesLeft = retriesLeft - 1
			time.Sleep(waitTime)
		} else {
			log.Printf("Connection successful!")
			break
		}

	}
	rand.Seed(time.Now().UnixNano())
	sid := rand.Intn(65534)

	serviceID := serviceName + "-" + strconv.Itoa(sid)

	consulService := api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Tags:    []string{time.Now().Format("Jan 02 15:04:05.000 MST")},
		Port:    port,
		Address: ip,
		Checks:  api.AgentServiceChecks{},
	}

	return srvRegistrator.ServiceRegister(&consulService)
}
