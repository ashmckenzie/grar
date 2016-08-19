#!/usr/bin/env ruby

require 'google/apis/calendar_v3'
require 'googleauth'
require 'googleauth/stores/file_token_store'
require 'active_support/all'

require 'fileutils'

OOB_URI = 'urn:ietf:wg:oauth:2.0:oob'
APPLICATION_NAME = 'Google Calendar API Ruby Quickstart'
CLIENT_SECRETS_PATH = 'client_secret.json'
CREDENTIALS_PATH = File.join(Dir.home, '.credentials', "calendar-ruby.json")
SCOPE = Google::Apis::CalendarV3::AUTH_CALENDAR_READONLY

##
# Ensure valid credentials, either by restoring from the saved credentials
# files or intitiating an OAuth2 authorization. If authorization is required,
# the user's default browser will be launched to approve the request.
#
# @return [Google::Auth::UserRefreshCredentials] OAuth2 credentials
def authorize
  FileUtils.mkdir_p(File.dirname(CREDENTIALS_PATH))

  client_id = Google::Auth::ClientId.from_file(CLIENT_SECRETS_PATH)
  token_store = Google::Auth::Stores::FileTokenStore.new(file: CREDENTIALS_PATH)
  authorizer = Google::Auth::UserAuthorizer.new(
    client_id, SCOPE, token_store)
  user_id = 'default'
  credentials = authorizer.get_credentials(user_id)
  if credentials.nil?
    url = authorizer.get_authorization_url(
      base_url: OOB_URI)
    puts "Open the following URL in the browser and enter the resulting code after authorization:\n\n"
    puts url
    print "\nCode: "
    code = gets
    puts
    credentials = authorizer.get_and_store_credentials_from_code(user_id: user_id, code: code, base_url: OOB_URI)
  end
  credentials
end

def convertsecs(x)
  h = x / 3600
  m = (x % 3600) / 60
  s = x % 60

  '%02d:%02d:%02d' % [ h, m, s ]
end

def i_accepted?(event)
  return false unless event.attendees
  event.attendees.each do |attendee|
    return true if attendee.email == 'amckenzie@zendesk.com' && attendee.response_status == 'accepted'
  end
  false
end

# Initialize the API
service = Google::Apis::CalendarV3::CalendarService.new
service.client_options.application_name = APPLICATION_NAME
service.authorization = authorize

# Fetch the next 10 events for the user
calendar_id = 'primary'
response = service.list_events(calendar_id,
                               single_events: true,
                               order_by: 'startTime',
                               time_min: Date.today.beginning_of_week.to_datetime.iso8601,
                               time_max: (Date.today.end_of_week - 1).to_datetime.iso8601
                              )

puts "* Meetings"

hours_in_working_day = 8
days_in_working_week = 5

seconds_in_a_working_day = (60 * 60) * hours_in_working_day
working_week_seconds = days_in_working_week * seconds_in_a_working_day
total_duration_seconds = 0

response.items.each do |event|
  next if !i_accepted?(event) || !event.end.date_time
  duration = event.end.date_time.strftime("%s").to_i - event.start.date_time.strftime("%s").to_i
  next if duration == 0
  total_duration_seconds += duration
  puts '  - %s (%s)' % [ event.summary, convertsecs(duration) ]
end

total_duration_seconds -= 20 * 60

puts "\nTotal %s (%.2f%% of week)" % [ convertsecs(total_duration_seconds), (total_duration_seconds.to_f / working_week_seconds.to_f) * 100 ]
