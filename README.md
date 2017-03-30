# Nick Bot

![](https://camo.githubusercontent.com/618e7a8fdb4d8b6b9bff3b9d2d70852bf40d6055/687474703a2f2f692e696d6775722e636f6d2f42646a6366734a2e676966)

## Dependencies

* OpenCV Developer package.

``` sh
$ sudo apt install libopencv-dev
```

## Built It

``` sh
$ git clone https://github.com/icholy/nick_bot.git
$ cd nick_bot
$ go get -d -v .
$ go build
```


## Usage

``` sh
Usage of ./nick_bot:
  -autofollow
    	auto follow random people
  -draw.face
    	Draw the face (default true)
  -draw.rects
    	Show the detection rectangles
  -face.dir string
    	directory to load faces from (default "faces")
  -face.opacity float
    	Face opacity [0-255] (default 1)
  -haar string
    	The location of the Haar Cascade XML configuration to be provided to OpenCV. (default "haarcascade_frontalface_alt.xml")
  -http.port string
    	http port (example :8080)
  -margin float
    	The face rectangle margin (default 60)
  -min.neighboor int
    	the lower this number is, the more faces will be found (default 9)
  -minfaces int
    	minimum faces (default 1)
  -password string
    	instagram password
  -post.interval duration
    	how often to post
  -post.now
    	post and exit
  -reset.store
    	mark all store records as available
  -sentry.dsn string
    	Sentry DSN
  -store string
    	the store file (default "store.db")
  -test.dir string
    	test a directory of images
  -test.image string
    	test image
  -upload
    	enable photo uploading
  -username string
    	instagram username
```

## Demo

![](https://raw.githubusercontent.com/icholy/nick_bot/master/demo.gif)
