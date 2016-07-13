package main

import (
  "os"
  "fmt"
  "time"

  "github.com/andygrunwald/go-jira"
  // "github.com/davecgh/go-spew/spew"
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
      "%s %s: %+v\n",
      t.In(localTZ).Format("2006-01-02T15:04:05-0700"),
      issue.Key,
      issue.Fields.Summary,
    )
  }
}

func getTickets(jiraDomain string, jiraUser string, jiraPassword string) {
  jiraClient, err := jira.NewClient(nil, jiraDomain)
  if err != nil {
    panic(err)
  }

  res, err := jiraClient.Authentication.AcquireSessionCookie(jiraUser, jiraPassword)
  if err != nil || res == false {
    fmt.Printf("Result: %v\n", res)
    panic(err)
  }

  // issue, _, err := jiraClient.Issue.Get("OP-20977")
  // if err != nil {
  //   panic(err)
  // }

  // fmt.Printf("https://zendesk.atlassian.net/browse/%s - %+v\n", issue.Key, issue.Fields.Summary)

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

func main() {
  var jiraDomain string
  var jiraUser string
  var jiraPassword string

  app := cli.NewApp()
  app.Name = "grar"
  app.Usage = "Jira command line tool"
  app.Version = "0.1.0"

  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "domain, d",
      Usage: "domain",
      EnvVar: "JIRA_DOMAIN",
      Destination: &jiraDomain,
    },
    cli.StringFlag{
      Name: "user, u",
      Usage: "user",
      EnvVar: "JIRA_USER",
      Destination: &jiraUser,
    },
    cli.StringFlag{
      Name: "password, p",
      Usage: "password",
      EnvVar: "JIRA_PASSWORD",
      Destination: &jiraPassword,
    },
  }

  app.Action = func(c *cli.Context) error {
    if len(jiraDomain) == 0 { return cli.NewExitError("ERROR: Domain is missing!", 1) }
    if len(jiraUser) == 0 { return cli.NewExitError("ERROR: User is missing!", 2) }
    if len(jiraPassword) == 0 { return cli.NewExitError("ERROR: Password is missing!", 3) }

    getTickets(jiraDomain, jiraUser, jiraPassword)
    return nil
  }

  app.Run(os.Args)
}
