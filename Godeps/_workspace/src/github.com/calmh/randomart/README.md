randomart [![Build Status](https://drone.io/github.com/calmh/randomart/status.png)](https://drone.io/github.com/calmh/randomart/latest)
=========

Generate OpenSSH style "randomart" images based on key fingerprints.

Documentation
-------------

http://godoc.org/github.com/calmh/randomart

Example
-------

```go
data := []byte{ 0x9b, 0x4c, 0x7b, 0xce, 0x7a, 0xbd, 0x0a, 0x13,
                0x61, 0xfb, 0x17, 0xc2, 0x06, 0x12, 0x0c, 0xed }
ra := randomart.Generate(data, "RSA 2048")
fmt.Println(ra)
```

```
+--[ RSA 2048 ]---+
|    .+.          |
|      o.         |
|     .. +        |
|      Eo =       |
|        S + .    |
|       o B . .   |
|        B o..    |
|         *...    |
|        .o+...   |
+-----------------+
```

License
-------

MIT
