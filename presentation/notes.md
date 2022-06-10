# Browser Tabs Open

- [docs.gofiber.io](https://docs.gofiber.io/)
- [gorm.io/docs](https://gorm.io/docs/)
- [npmjs.com/package/eyeson](https://www.npmjs.com/package/eyeson)

# Workshop Step 0

Basic project setup, start a minimal server and test fiber locally.

```sh
$ git init
$ go mod init goose
$ mkdir cmd
$ vim cmd/server.go
```

See how fiber works.

```go
// $ go get github.com/gofiber/fiber/v2
// cmd/server.go
package main
import "github.com/gofiber/fiber/v2"

func main() {
	port := 8077

	app := fiber.New(fiber.Config{})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello world\n")
	})

	fmt.Printf("Start server on port %d\n", port)
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal(err)
	}
}
```

# Workshop Step 1

Create workspaces and start meetings.

- switch to step-1 directory, open the README
- lets start with our database model
- then have a look on the routes `goose.go`, and `cmd/server.go`
- at last, check controllers and test the application
- implement the meetings action(!)

```go
func MeetingsCreate(c *fiber.Ctx) error {
  id := c.Params("workspace_id")
  options := map[string]string{
    "options[sfu_mode]": "disabled",
    "options[exit_url]": c.BaseURL() + "/",
  }
  meeting, err := videoService.Rooms.Join(id, "user", options)
  if err != nil {
    log.Println(err)
    return fiber.ErrBadRequest
  }
  return c.Redirect(meeting.Data.Links.Gui, 303)
}
```

- conclude: we can already start independent meetings grouped by our
  workspaces

- `find . -name "*.go" | xargs wc -l`, including newlines and comments it took
  about 120 lines of code to solve our first task.

# Workshop Step 2.

Add users, cookies, logins.

- show added user and login model
- discuss new routes as overview, comment them
- implement create NewUser method, extract name from email

```go
func NewUser(email string) (*User, error) {
	user := User{Email: email, Name: ExtractNameFromEmail(email)}
	result := db.Where(&User{Email: email}).FirstOrCreate(&user)
	return &user, result.Error
}

// ExtractNameFromEmail extracts a name from a given company email like,
// christoph.lipautz@eyeson.com => christoph.lipautz => christoph lipautz
func ExtractNameFromEmail(email string) string {
	name := strings.ToLower(strings.Split(email, "@")[0])
	return strings.ReplaceAll(name, ".", " ")
}
```

- show login view, instructions page, new routes, handle root path
- logins create: check email domain HasSuffix, create auth (generate), send email
- logins show: take auth, check with db, set cookie, redirect to root path
- add authenticated helper, take username for meeting


- lets take it a step further and don't leave our platform for the video
  meetings.
- create `meetings#show`

```go
log.Println("Meeting Guest Join:", room.Data.Links.GuestJoin)
return c.Render("meetings/show", fiber.Map{"AccessKey": room.Data.AccessKey})
```

```html
// log GuestLink in meetings show so we can join with another user.
<a href="/">Back to Workspaces</a>

<div id="meeting-container">
  <video id="meeting-stream" data-access-key="{{.AccessKey}}" autoplay></video>
</div>

<script src="/assets/eyeson.js"></script>
<script>
const eyeson = window.eyeson.default;

const video = document.querySelector("video");

eyeson.onEvent(event => {
  if (event.type !== "accept") {
    console.debug("event received", event.type);
    return;
  }
  video.srcObject = event.remoteStream;
  video.play();
});
eyeson.start(video.dataset.accessKey);
</script>
```

# Workshop Step 3.

- show models for meeting and recordings
- show webhook registration
- show debug output for webhook

```go
debug, _ := json.Marshal(webhook)
fmt.Printf("Received webhook of type %v, with content:\n%v\n", webhook.Type, debug)
```

- test webhook received

# Workshop The Full Picture

- shortly explain the additional work done
- demo and show snapshots, recordings, meetings
- show markdown content for workspaces and meetings

# Workshop Step 4, Final Picture

- demo final application
- show meetings view
- show snapshots handling
