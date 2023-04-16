# Skeleton - Basic HTTP Server setup

```go
import "github.com/cyc-ttn/skeleton"
```

Skeleton wraps the provided HTTP server with a set of interfaces for 
applications to provide their own implementations for session management and 
routing. 

It includes the general code necessary to:
- handle graceful shutdown 
- allow long-standing routines to be connected to said graceful shutdown

Optionally, it also provides a way to add an application's own logging to a 
service. In addition to the above, the logging HTTP server will:
- create a unique request ID for each request and store it in the header
- create a request logger for each request 
- start the timer on the request logger. 

For example, see the [examples](/examples) directory. 

## License

Skeleton is released under the [MIT License](http://www.opensource.org/licenses/MIT).