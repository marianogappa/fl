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

### Caveats/Disclaimers

- Loading sqlite3 dump

### TODO

- TODO: unit test readCSV?
- TODO: search many fields: include urls in full-text search
