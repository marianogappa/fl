## fl

### Test

```
# Requires docker, docker-compose
$ make test
```

### Run

```
# Requires go, docker, docker-compose
$ make run
```

### Design decisions

- Elasticsearch: industry standard for search; full-text + geo_point out-of-the-box
- Elasticsearch cluster: high availability and horizontal scaling as data grows and request volume grows
- Go µs endpoint is stateless: n load balanced replicas for high availability and horizontal scaling
- Testing: integration tests as there is nothing to test other than ES query + endpoint contract
- Please refer to [db.go](db.go)'s search function for a detailed explanation of how results are chosen and sorted

### Caveats/Disclaimers

- Used this line to dump the sqlite db into a csv: `sqlite3 -csv fatlama.sqlite3 "SELECT * FROM items" > dump.csv`
- I didn't use an ES cluster (i.e. only one replica) or load balancing for the µs, but both are designed for it
- [Here's](example-kubernetes-deployment.yml) an example Kubernetes service/deploy for the µs with 5 replicas
- Similar HA options are available for Elasticsearch, being in k8s, Mesos, docker itself or bare bones
