
# Create a Video Platform From Scratch

A minimalistic (web) video platform to organize meetings between users.

As a company we want to avoid 3rd-party tools for internal video communication
and therefor use a custom internal video platform to host our internal remote
meetings.

The application ensures the following requirements are met:

- Authenticate Users
- Start and Join a Meeting
- Show Meeting Information, including Download of Snapshots and Recordings
- Use a Basic Custom Video User Interface

With the encapsulated business logic we can show that the main platform parts
can easily be used to be part of any existing platform or software application.

Work Project Name: MeetDuck - Producktive Video Meetings

Go Module Name: goose

## Usage & Development

Start a local server on port 8077.

```sh
$ make server API_KEY=... WH_URL=...
$ make test # run testsuite
$ make watch # watch for changes and run tests
$ npm pack eyeson # fetch eyeson JavaScript package
```

Build and run a container image.

```sh
$ make build-image
$ make run-image
```

## Specifications

MeetDuck is a intranet like video platform. The software is registered within
a company domain, e.g. eyeson.com, that allows any email address of this domain
to authenticate. The authentication process is handled by sending an email with
a unique short lived authentication link.

A user can create a workspace - that is a topic based organizational unit - and
assign others to be part of it. Within a workspace, people can start and join
meetings.

A workspace stores the history of past meetings, recordings and snapshots.

## Data Schema

For sake of simplicity we make avoid moderator roles and allow any user to
manage content as long as they are associated with the given workspace.

```yaml
User:
  - Name
  - Email
  - Workspaces

Login:
  - AuthKey
  - ExpiesAt
  - User

Workspace:
  - Topic
  - Content

Meeting:
  - StartedAt
  - EndedAt
  - Workspace

Recording:
  - Reference
  - Duration
  - StartedAt
  - EndedAt
  - Workspace

Snapshots:
  - Workspace
```

## Routes

```
GET / ... login (main entrance point)
POST /sessions ... create a new user login (send email w/ login link)
GET /sessions/:auth ... create a new user session

# -- authenticated --
GET / ... workspaces overview
GET /workspaces/new ... new workspace form
POST /workspaces ... create a new workspace
GET /workspaces/:id ... show a workspace, including events
GET /workspaces/:id/edit ... show a workspace update form
POST /workspaces/:id/edit ... update a workspace
POST /workspaces/:id/users ... assign a user to workspace
POST /workspace/:id/meeting ... join a workspace meeting

GET /meetings/:id/edit ... show edit form of a meeting event
POST /meetings/:id/edit ... update a meeting event

GET /recordings/:id.webm ... fetch a recording
GET /snapshots/:id.webm ... fetch a snapshot

GET /signout ... delete a session

POST /webhook ... receive a webhook
```

## References

- [eyeson Go](https://github.com/eyeson-team/eyeson-go)
- [eyeson API docs](https://eyeson-team.github.io/api/api-reference/)
- [eyeson API key](https://developers.eyeson.team/)
- [eyeson JS](https://www.npmjs.com/package/eyeson)
- [eyeson JS docs](https://eyeson-team.github.io/js-docs/overview/)
- [fiber docs](https://docs.gofiber.io/)
- [gorm docs](https://gorm.io/docs/)
- [entr file watcher](https://github.com/eradman/entr)
