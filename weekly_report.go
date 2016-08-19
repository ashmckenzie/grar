package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"
  "time"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/calendar/v3"
  googleOAuth2 "google.golang.org/api/oauth2/v2"
  "github.com/jinzhu/now"
  "github.com/ashmckenzie/grar/secrets"
  "github.com/urfave/cli"
  // "github.com/davecgh/go-spew/spew"
)

type User struct {
  Sub string `json:"sub"`
  Name string `json:"name"`
  GivenName string `json:"given_name"`
  FamilyName string `json:"family_name"`
  Profile string `json:"profile"`
  Picture string `json:"picture"`
  Email string `json:"email"`
  EmailVerified string `json:"email_verified"`
  Gender string `json:"gender"`
}

var appSecrets = secrets.NewAppSecrets()
var me User

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
  cacheFile, err := tokenCacheFile()
  if err != nil { log.Fatalf("Unable to get path to cached credential file. %v", err) }

  tok, err := tokenFromFile(cacheFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(cacheFile, tok)
  }

  return config.Client(ctx, tok)
}

func setUser(client *http.Client) {
  resp, _ := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
  defer resp.Body.Close()
  data, _ := ioutil.ReadAll(resp.Body)
  json.Unmarshal(data, &me)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  var code string

  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Printf("Go to the following link in your browser then type the authorization code: \n\n%v\n\nCode: ", authURL)

  if _, err := fmt.Scan(&code); err != nil { log.Fatalf("Unable to read authorization code %v", err) }
  fmt.Printf("\n")

  tok, err := config.Exchange(oauth2.NoContext, code)
  if err != nil { log.Fatalf("Unable to retrieve token from web %v", err) }

  return tok
}

func tokenCacheFile() (string, error) {
  usr, err := user.Current()
  if err != nil { return "", err }

  tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
  os.MkdirAll(tokenCacheDir, 0700)

  return filepath.Join(tokenCacheDir, url.QueryEscape("calendar-go.json")), err
}

func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil { return nil, err }

  t := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(t)
  defer f.Close()

  return t, err
}

func saveToken(file string, token *oauth2.Token) {
  f, err := os.Create(file)
  if err != nil { log.Fatalf("Unable to cache oauth token: %v", err) }
  defer f.Close()
  json.NewEncoder(f).Encode(token)
}

func convertSecs(x int) (string) {
  h := x / 3600
  m := (x % 3600) / 60
  s := x % 60

  return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func timeIntoSeconds(x string) (int) {
  t, err := time.Parse("2006-01-02T15:04:05-07:00", x)
  if err != nil { log.Fatalf("Unable to parse time %v", err) }

  return int(t.Unix())
}

func iAccepted(x *calendar.Event) (bool) {
  for _, i := range x.Attendees {
    if (i.Email == me.Email && i.ResponseStatus == "accepted") { return true }
  }

  return false
}

func generateReport() {
  ctx := context.Background()

  config, err := google.ConfigFromJSON([]byte(appSecrets.ClientJSON),
                                              calendar.CalendarReadonlyScope,
                                              googleOAuth2.UserinfoEmailScope)
  if err != nil { log.Fatalf("Unable to parse client secret file to config: %v", err) }

  client := getClient(ctx, config)
  setUser(client)

  srv, err := calendar.New(client)
  if err != nil { log.Fatalf("Unable to retrieve calendar Client %v", err) }

  timeMin := now.BeginningOfWeek().Format(time.RFC3339)
  timeMax := now.BeginningOfWeek().AddDate(0, 0, 6).Add(-time.Nanosecond).Format(time.RFC3339)

  events, err := srv.Events.
                      List("primary").
                      ShowDeleted(false).
                      SingleEvents(true).
                      TimeMin(timeMin).
                      TimeMax(timeMax).
                      OrderBy("startTime").
                      MaxResults(500).
                      Do()

  if err != nil { log.Fatalf("Unable to retrieve next ten of the user's events. %v", err) }

  fmt.Println("* Meetings")

  if len(events.Items) > 0 {
    totalDuration := 0
    now.FirstDayMonday = true

    for _, i := range events.Items {
      if (i.End.DateTime == "" || i.Start.DateTime == "" || !iAccepted(i)) { continue }

      startTimeDuration := timeIntoSeconds(i.Start.DateTime)
      endTimeDuration := timeIntoSeconds(i.End.DateTime)

      duration := endTimeDuration - startTimeDuration
      if (duration == 0) { continue }

      totalDuration += duration

      fmt.Printf("  - %s (%s)\n", i.Summary, convertSecs(duration))
    }

    secondsInAWorkingDay := (60 * 60) * 8
    workingWeekSeconds := 5 * secondsInAWorkingDay

    fmt.Printf("\nTotal %s (%.2f%% of week)", convertSecs(totalDuration), (float32(totalDuration) / float32(workingWeekSeconds)) * 100)
  }
}

func main() {
  app := cli.NewApp()
  app.Name = "weekly_report"
  app.Usage = "Weekly Report using Google Calendar"
  app.Version = "0.1.1"

  app.Action = func(c *cli.Context) error {
    generateReport()

    return nil
  }

  app.Run(os.Args)
}
