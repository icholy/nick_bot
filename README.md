# Nick Bot

> Adding some much needed nick to your Instagram photos.

![](https://camo.githubusercontent.com/618e7a8fdb4d8b6b9bff3b9d2d70852bf40d6055/687474703a2f2f692e696d6775722e636f6d2f42646a6366734a2e676966)

## Dependencies

* OpenCV Developer package.

``` sh
$ sudo apt install libopencv-dev
```

## Build

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

## Design

### Posting Schedule:

> The bot posts at [peak engagement times](http://www.huffingtonpost.com/2015/02/25/get-instagram-likes_n_6751614.html)

* Sunday: 6:00AM and 5:00PM
* Monday: 2:00AM and 7:00PM
* Tuesday: 3:00AM and 10:00PM
* Wednesday: 2:00AM and 5:00PM
* Thursday: 7:00AM and 5:00PM
* Friday: 1:00AM and 8:00PM
* Saturday: 2:00AM and 7:00PM

The schedule is configured using the `schedule.cron` file (uses cron format).

### Auto-Follow

> The bot automatically follows people.

* The bot needs a consistent flow of new followers to find new pictures.
* When someone's photo is modified/posted, the bot automatically follows 1-10 of their followers.

### Image Selection

> When it's time to post, an photo must be selected from the image store.

* There are several image selection 'strategies'.
* Each strategy has a different probability of being chosen.

##### Strategies:

1. Top faces of all followers (P: 0.40)
2. Top likes of all followers (P: 0.40)
3. Most faces of a random follower (P: 0.10)
4. Most likes of a random follower (P: 0.10)

* **Note**: score is `likes * faces`.

### Crawler

> The crawler's job is to find follower's photos.

* Every 0-30 minutes it downloads a list of all followers and shuffles them.
* Every 1-60 seconds it downloads one photo from a user.
* After a photo is downloaded, the faces are detected, and the metadata written to the store.

### Image Store

> The image store is an index of all crawled photos.

* [SQLite](https://www.sqlite.org/) database with the following structure:

``` sql
CREATE TABLE media (
  media_id    TEXT,    -- photo id
  media_url   TEXT,    -- photo url
  user_id     TEXT,    -- original poster user id
  user_name   TEXT,    -- original poster username
  like_count  INTEGER, -- number of likes
  face_count  INTEGER, -- number of detected faces
  posted_at   INTEGER, -- timestamp of when the original was posted
  state       INTEGER  -- available, used, or rejected
);
```

### Face Detection

> Uses a Haar Feature-based Cascade Classifier for Object Detection.

* http://docs.opencv.org/2.4/modules/objdetect/doc/cascade_classification.html

## Demo

![](https://raw.githubusercontent.com/icholy/nick_bot/master/demo.gif)

## Relevant Instagram TOS

* Share only photos and videos that you’ve taken or have the right to share.

> Remember to post authentic content, and don’t post anything you’ve copied or collected from the Internet that you don’t have the right to post.

* Foster meaningful and genuine interactions.

> Help us stay spam-free by not artificially collecting likes, followers, or shares, posting repetitive comments or content, or repeatedly contacting people for commercial purposes without their consent.

* Respect other members of the Instagram community.

> We remove [...] content that targets private individuals to degrade or shame them, personal information meant to blackmail or harass someone, and repeated unwanted messages.
