package main

import (
  "os"
  // "encoding/json"
  "fmt"
  "time"

  "github.com/andygrunwald/go-jira"
  // "github.com/davecgh/go-spew/spew"
  //"github.com/jroimartin/gocui"
  "github.com/urfave/cli"
)

func getIssues(jiraClient *jira.Client, opt *jira.SearchOptions, query string) ([]jira.Issue, *jira.Response) {
  issues, response, _ := jiraClient.Issue.Search(query, opt)
  return issues, response
}

func printIssues(localTZ *time.Location, issues []jira.Issue) {
  for _, issue := range issues {
    t, _ := time.Parse("2006-01-02T15:04:05.000-0700", issue.Fields.Updated)
    fmt.Printf(
      //"%s %s (https://zendesk.atlassian.net/browse/%s): %+v\n",
      "%s %s: %+v\n",
      t.In(localTZ).Format("2006-01-02T15:04:05-0700"),
      issue.Key,
      //issue.Key,
      issue.Fields.Summary,
    )
  }
}

func jiraConnect(jiraDomain string, jiraUser string, jiraPassword string) (*jira.Client) {
  jiraClient, err := jira.NewClient(nil, jiraDomain)
  if err != nil { panic(err) }

  res, err := jiraClient.Authentication.AcquireSessionCookie(jiraUser, jiraPassword)
  if err != nil || res == false {
    fmt.Printf("Result: %v\n", res)
    panic(err)
  }

  return jiraClient
}

func getTickets(jiraClient *jira.Client) {
  localTZ, _ := time.LoadLocation("Australia/Melbourne")

  resolvedTicketsQuery := "assignee = currentUser() AND status in (Resolved, Closed) AND resolutiondate >= startOfWeek() ORDER BY resolutiondate ASC"
  updatedTicketsQuery := "assignee = currentUser() AND status not in (Resolved, Closed) AND updatedDate >= startOfWeek() ORDER BY updatedDate ASC"

  opt := &jira.SearchOptions{StartAt: 0, MaxResults: 40}

  issues, response := getIssues(jiraClient, opt, resolvedTicketsQuery)
  fmt.Printf("\nResolved tickets (%d)\n--------------------\n", response.Total)
  printIssues(localTZ, issues)

  issues, response = getIssues(jiraClient, opt, updatedTicketsQuery)
  fmt.Printf("\nUpdated tickets (%d)\n--------------------\n", response.Total)
  printIssues(localTZ, issues)
}

func wip(jiraClient *jira.Client) {
  // // opetrushka, amkenzie, dkertesz
  // issue, _, err := jiraClient.Issue.Get("OP-21266")
  // if err != nil { panic(err) }
  // // fmt.Printf("%s: %s (%s)\n", issue.Key, issue.Fields.Summary, issue.Fields.Assignee.Name)

  // for _, element := range issue.Fields.Subtasks {
  //   payloadStr := `{ "fields": { "assignee": { "name": "opetrushka" } } }`
  //   var payload map[string]interface{}
  //   json.Unmarshal([]byte(payloadStr), &payload)

  //   url := fmt.Sprintf("/rest/api/2/issue/%s", element.Key)
  //   req, _ := jiraClient.NewRequest("PUT", url, payload)
  //   _, err := jiraClient.Do(req, nil)
  //   if err != nil { panic(err) }
  // }
}

func main() {
  var jiraDomain string
  var jiraUser string
  var jiraPassword string
  var jiraClient *jira.Client

  app := cli.NewApp()
  app.Name = "grar"
  app.Usage = "Jira command line tool"
  app.Version = "0.1.0"

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "domain",
      Usage: "domain",
      EnvVar: "JIRA_DOMAIN",
      Destination: &jiraDomain,
    },
    cli.StringFlag{
      Name: "user",
      Usage: "user",
      EnvVar: "JIRA_USER",
      Destination: &jiraUser,
    },
    cli.StringFlag{
      Name: "password",
      Usage: "password",
      EnvVar: "JIRA_PASSWORD",
      Destination: &jiraPassword,
    },
  }

  app.Commands = []cli.Command{
    {
      Name: "report",

      Action: func(c *cli.Context) error {
        if len(jiraDomain) == 0 { return cli.NewExitError("ERROR: Domain is missing!", 1) }
        if len(jiraUser) == 0 { return cli.NewExitError("ERROR: User is missing!", 2) }
        if len(jiraPassword) == 0 { return cli.NewExitError("ERROR: Password is missing!", 3) }

        jiraClient = jiraConnect(jiraDomain, jiraUser, jiraPassword)
        getTickets(jiraClient)
        return nil
      },
    },
    {
      Name: "wip",
      Action: func(c *cli.Context) error {
        if len(jiraDomain) == 0 { return cli.NewExitError("ERROR: Domain is missing!", 1) }
        if len(jiraUser) == 0 { return cli.NewExitError("ERROR: User is missing!", 2) }
        if len(jiraPassword) == 0 { return cli.NewExitError("ERROR: Password is missing!", 3) }

        jiraClient = jiraConnect(jiraDomain, jiraUser, jiraPassword)
        wip(jiraClient)
        return nil
      },
    },
  }

  app.Run(os.Args)
}
