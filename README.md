byond-topic-go
==============

A library to send BYOND Topic messages
to Dream Daemon servers, and get a
string result.

This only expects the string result from a
BYOND Dream Daemon server - if you have a
case for the other type of result (floating point),
please make an issue above.

Using
-----

``` go
import "github.com/vulppine/byond-topic-go"

func main() {
    r := byondtopic.SendTopic("[dream daemon address:port]", "[topic]")
    // do something with r
}
```

License
-------

Flipp Syder, MIT License, 2021
    
         
