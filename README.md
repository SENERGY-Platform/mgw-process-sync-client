synchronises process states of the local camunda instance to a cloud service by communicating with a github.com/SENERGY-Platform/process-sync service via mqtt 


## OpenAPI
uses https://github.com/swaggo/swag

### installation
```
go install github.com/swaggo/swag/cmd/swag@latest
```

### generating
```
swag init --parseDependency -d ./pkg/events/api -g api.go
```

### swagger ui
if the config variable UseSwaggerEndpoints is set to true, a swagger ui is accessible on /swagger/index.html (http://localhost:8080/swagger/index.html)