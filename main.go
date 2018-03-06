package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/monochromegane/go-gitignore"
	"github.com/ryanuber/columnize"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type FileJob struct {
	Language  string
	Filename  string
	Extension string
	Location  string
	Content   []byte
	Bytes     int64
	Lines     int64
	Code      int64
	Comment   int64
	Blank     int64
}

type LanguageSummary struct {
	Name    string
	Bytes   int64
	Lines   int64
	Code    int64
	Comment int64
	Blank   int64
	Count   int64
}

const (
	database_languages = `Ww0KICB7DQogICAgImxhbmd1YWdlIjogIlRleHQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInRleHQiLA0KICAgICAgInR4dCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiWEFNTCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAieGFtbCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQVNQLk5ldCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiYXNjeCIsDQogICAgICAiYXNteCIsDQogICAgICAiYXNheCIsDQogICAgICAiYXNweCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSFRNTCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiaHRtIiwNCiAgICAgICJodG1sIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJNU0J1aWxkIHNjcmlwdHMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImNzcHJvaiINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQyMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImNzIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJYU0QiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInhzZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiWE1MIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ4bWwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkNNYWtlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjbWFrZSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQy9DKytIZWFkZXIiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImhwcCIsDQogICAgICAiaHh4IiwNCiAgICAgICJoaCIsDQogICAgICAiaW5sIiwNCiAgICAgICJpcHAiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkMrKyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiY3BwIiwNCiAgICAgICJjYyIsDQogICAgICAiY3h4Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJtYWtlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJtYWtlZmlsZSIsDQogICAgICAiZ251bWFrZWZpbGUiLA0KICAgICAgIm1ha2UiLA0KICAgICAgIm1rIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJDU1MiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImNzcyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUHl0aG9uIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJweSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTUFUTEFCIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJtIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJPYmplY3RpdmVDIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJtIiwNCiAgICAgICJtbSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSmF2YXNjcmlwdCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAianMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkphdmEiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImphdmEiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlBIUCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicGhwNCIsDQogICAgICAicGhwNSIsDQogICAgICAicGhwIiwNCiAgICAgICJpbmMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkVybGFuZyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZXJsIiwNCiAgICAgICJocmwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkZvcnRyYW4gNzciLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImYiLA0KICAgICAgImY3NyIsDQogICAgICAiZm9yIiwNCiAgICAgICJmdG4iLA0KICAgICAgImZwcCIsDQogICAgICAiZjk1IiwNCiAgICAgICJmMDMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkZvcnRyYW4gOTAiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImY5MCIsDQogICAgICAiZiIsDQogICAgICAiZm9yIiwNCiAgICAgICJmdG4iLA0KICAgICAgImZwcCIsDQogICAgICAiZjk1IiwNCiAgICAgICJmMDMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImMiLA0KICAgICAgImVjIiwNCiAgICAgICJwZ2MiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkxpc3AiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImxpc3AiLA0KICAgICAgImVsIiwNCiAgICAgICJsc3AiLA0KICAgICAgInNjIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJWaXN1YWwgQmFzaWMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInZicyIsDQogICAgICAidmIiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkJvdXJuZSBTaGVsbCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAic2giLA0KICAgICAgImJhc2giLA0KICAgICAgInpzaCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUnVieSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicmIiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogInZpbXNjcmlwdCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAidmltIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJBc3NlbWJseSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicyIsDQogICAgICAiYXNtIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJPYmplY3RpdmUgQysrIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJtbSIsDQogICAgICAibSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiRFREIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJkdGQiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlNRTCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAic3FsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJZQU1MIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ5YW1sIiwNCiAgICAgICJ5bWwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlJ1YnkgSFRNTCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicmh0bWwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkhhc2tlbGwiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImhzIiwNCiAgICAgICJsaHMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkJvdXJuZSBBZ2FpbiBTaGVsbCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAic2giLA0KICAgICAgImJhc2giLA0KICAgICAgInpzaCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQWN0aW9uU2NyaXB0IiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJhcyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTVhNTCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAibXhtbCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQVNQIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJhc2EiLA0KICAgICAgImFzcCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiRCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUGFzY2FsIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJwIiwNCiAgICAgICJwYXMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlNjYWxhIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJzY2FsYSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiRE9TQmF0Y2giLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImJhdCIsDQogICAgICAiY21kIiwNCiAgICAgICJiYXQiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkdyb292eSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZ3Jvb3Z5Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJYU0xUIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ4c2wiLA0KICAgICAgInhzbHQiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlBlcmwiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImNnaSIsDQogICAgICAicGwiLA0KICAgICAgInBtIiwNCiAgICAgICJwbTYiLA0KICAgICAgInBvZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiVGVhbWNlbnRlciBkZWYiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImRlZiINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSURMIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJpZGwiLA0KICAgICAgInBybyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTHVhIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJsdWEiLA0KICAgICAgInJvY2tzcGVjIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJHbyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZ28iDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogInlhY2MiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInkiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkN5dGhvbiIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicHl4Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJsZXgiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkFkYSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicGFkIiwNCiAgICAgICJhZGIiLA0KICAgICAgImFkcyIsDQogICAgICAiYWRhIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJzZWQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInNlZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAibTQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgIm00IiwNCiAgICAgICJhYyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiT2NhbWwiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgIm1sIiwNCiAgICAgICJtbGkiLA0KICAgICAgIm1seSIsDQogICAgICAibWxsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJTbWFydHkiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInRwbCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQ29sZEZ1c2lvbiIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiY2ZtIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJOQW50IHNjcmlwdHMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImJ1aWxkIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJFeHBlY3QiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImV4cCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQ1NoZWxsIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJzaCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiVkhETCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAidmhkIiwNCiAgICAgICJ2aGRsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJUY2wvVGsiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInRjbCIsDQogICAgICAidGsiLA0KICAgICAgInRrcGtnIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJKU1AiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImpzcCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiU0tJTEwiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImlsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJhd2siLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImF3ayINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTVVNUFMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgIm0iLA0KICAgICAgIm1wcyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiS29ybiBTaGVsbCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAia3NoIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJGb3J0cmFuIDk1IiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJmOTUiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIk9yYWNsZSBGb3JtcyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZm10Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJEYXJ0IiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJkYXJ0Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJDT0JPTCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiY29iIiwNCiAgICAgICJjYmwiLA0KICAgICAgImNibCIsDQogICAgICAiY29iIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJNb2R1bGEzIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJpMyIsDQogICAgICAibWciLA0KICAgICAgImlnIiwNCiAgICAgICJtMyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiT3JhY2xlIFJlcG9ydHMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInJleCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiU29mdGJyaWRnZSBCYXNpYyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAic2JsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJNYXJrZG93biIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAibWQiLA0KICAgICAgIm1hcmtkb3duIiwNCiAgICAgICJtZG93biIsDQogICAgICAibWR3biIsDQogICAgICAibWtkbiIsDQogICAgICAibWtkIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJnaXQtaWdub3JlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJnaXRpZ25vcmUiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkZyZWVtYXJrZXIgVGVtcGxhdGUiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImZ0bCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTGVzc0NTUyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAibGVzcyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiR3JhZGxlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJncmFkbGUiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkJhc2ljIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJiYXMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkdvbGZTY3JpcHQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImdvbGZzY3JpcHQiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkxhVGVYIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ0ZXgiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkJvbyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiYm9vIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJKdWxpYSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiamwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkRlbHBoaSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZGVscGhpIiwNCiAgICAgICJwYXMiLA0KICAgICAgImRmbSIsDQogICAgICAibmZtIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJMT0xDT0RFIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJsb2wiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkIiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImIiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkNoZWYiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImNoIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJSYWNrZXQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInJrdCIsDQogICAgICAicmt0bCIsDQogICAgICAic3MiLA0KICAgICAgInNjbSIsDQogICAgICAic2NyYmwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlN3aWZ0IiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJzd2lmdCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSlNPTiIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAianNvbiINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiT2N0YXZlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJvY3RhdmUiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkVsaXhpciIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZXhzIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJGYWN0b3IiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImZhY3RvciINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiVmltU2NyaXB0IiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ2aW0iDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlBvd2Vyc2hlbGwiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInBzMSIsDQogICAgICAicHNtMSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiRWlmZmVsIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJlaWZmIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJTY2FsYWJsZSBWZWN0b3IgR3JhcGhpY3MiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInN2ZyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUnVzdCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicnMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIk1VU0hDb2RlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJtdXNoIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJMb2dvIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJsZyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTmltIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJuaW0iDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIldvbGZyYW0gTGFuZ3VhZ2UiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgIndsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJQdXJlYmFzaWMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInBiIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJBcm5vbGRDIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJhcm5vbGRjIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJWUk1MIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ3cmwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkNvZmZlZXNjcmlwdCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiY29mZmVlIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJTUERYIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJzcGR4Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJUeXBlU2NyaXB0IiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJ0cyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSlNYIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJqc3giDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlJ1YnkgVGVtcGxhdGUiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImVyYiINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiWE1MIFJlc291cmNlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJyZXN4Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJWYXJuaXNoIENvbmZpZ3VyYXRpb24iLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInZjbCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSmFkZSBUZW1wbGF0ZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiamFkZSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUmVTdHJ1Y3R1cmVkIFRleHQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInJzdCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQ1NWIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjc3YiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlJhem9yIFRlbXBsYXRlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjc2h0bWwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkhhbmRsZWJhcnMgVGVtcGxhdGUiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImhicyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSU5JRmlsZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiaW5pIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJDb25maWd1cmF0aW9uIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjb25mIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJDbG9qdXJlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjbGoiLA0KICAgICAgImNsanMiLA0KICAgICAgImNsamMiLA0KICAgICAgImNsangiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlZpc3VhbCBORGVwZW5kIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJuZHByb2oiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkRldmljZSBUcmVlIFNvdXJjZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZHRzIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJBU1AuTkVUIFdlYkhhbmRsZXIiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImFzaHgiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkdoZXJraW4gU3BlY2lmaWNhdGlvbiIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZmVhdHVyZSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSGF4ZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiaHgiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlF0IE1ldGEgTGFuZ3VhZ2UiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInFtbCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiU3R5bGUgU2hlZXQgZVh0ZW5kZXIiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImNzc3giDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlNjcmF0Y2ggUHJvamVjdCBGaWxlIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJzYiINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiT3BhbGFuZyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAib3BhIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJQb3J0YWdlIEluc3RhbGxlciIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZWJ1aWxkIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJDcnlzdGFsIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjciINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiSmVua2lucyBCdWlsZGZpbGUiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImplbmtpbnNmaWxlIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJLb3RsaW4iLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImt0Ig0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJQcm9wZXJ0aWVzIEZpbGUiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInByb3BlcnRpZXMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlRhZyBMaWJyYXJ5IERlc2NyaXB0b3IiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInRsZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiU3ludGFjdGljYWxseSBBd2Vzb21lIFN0eWxlIFNoZWV0cyIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAic2NzcyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiRW1iZWRkZWQgSmF2YVNjcmlwdCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZWpzIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJQYXRjaCIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicGF0Y2giDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkNvY29hcG9kIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJjb2NvYXBvZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiRiMiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImZzIiwNCiAgICAgICJmc2kiLA0KICAgICAgImZzeCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTWF0aGVtYXRpY2EiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgIm0iLA0KICAgICAgIndsIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJQYXJyb3QiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInBpciIsDQogICAgICAicGFzbSIsDQogICAgICAicG1jIiwNCiAgICAgICJvcHMiLA0KICAgICAgInBvZCIsDQogICAgICAicGciLA0KICAgICAgInRnIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJQdXBwZXQiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInBwIg0KICAgIF0NCiAgfSwNCiAgew0KICAgICJsYW5ndWFnZSI6ICJSYWtlZmlsZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAicmFrZWZpbGUiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlBLR0JVSUxEIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJwa2didWlsZCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiVG9tJ3MgT2J2aW91cywgTWluaW1hbCBMYW5ndWFnZSAiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInRvbWwiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIkxvY2sgRmlsZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAibG9jayINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiTGljZW5zZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAibGljZW5zZSIsDQogICAgICAiY29weWluZyIsDQogICAgICAiY29weWluZzMiDQogICAgXQ0KICB9LA0KICB7DQogICAgImxhbmd1YWdlIjogIlR5cGluZ3MgRGVmaW5pdGlvbiIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiZC50cyINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUm9ib3QgRnJhbWV3b3JrIiwNCiAgICAiZXh0ZW5zaW9ucyI6IFsNCiAgICAgICJyb2JvdCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiUmVwb3J0IERlZmluaXRpb24gTGFuZ3VhZ2UiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgInJkbCINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiS2V5IEZpbGUiLA0KICAgICJleHRlbnNpb25zIjogWw0KICAgICAgImtleSINCiAgICBdDQogIH0sDQogIHsNCiAgICAibGFuZ3VhZ2UiOiAiQ2VydGlmaWNhdGUgRmlsZSIsDQogICAgImV4dGVuc2lvbnMiOiBbDQogICAgICAiY3J0Ig0KICAgIF0NCiAgfQ0KXQ==`
)

var Exclusions = strings.Split("woff,eot,cur,dm,xpm,emz,db,scc,idx,mpp,dot,pspimage,stl,dml,wmf,rvm,resources,tlb,docx,doc,xls,xlsx,ppt,pptx,msg,vsd,chm,fm,book,dgn,blines,cab,lib,obj,jar,pdb,dll,bin,out,elf,so,msi,nupkg,pyc,ttf,woff2,jpg,jpeg,png,gif,bmp,psd,tif,tiff,yuv,ico,xls,xlsx,pdb,pdf,apk,com,exe,bz2,7z,tgz,rar,gz,zip,zipx,tar,rpm,bin,dmg,iso,vcd,mp3,flac,wma,wav,mid,m4a,3gp,flv,mov,mp4,mpg,rm,wmv,avi,m4v,sqlite,class,rlib,ncb,suo,opt,o,os,pch,pbm,pnm,ppm,pyd,pyo,raw,uyv,uyvy,xlsm,swf", ",")
var ExtensionToLanguage = map[string]string{"scm": "Scheme", "asmx": "ASP.NET", "handlebars": "Handlebars", "less": "LESS", "csproj": "MSBuild", "hamlet": "Hamlet", "lds": "LD Script", "tcl": "TCL", "gd": "GDScript", "lsp": "Lisp", "go": "Go", "sv": "SystemVerilog", "xml": "XML", "ads": "Ada", "clj": "Clojure", "elm": "Elm", "ihex": "Intel HEX", "ede": "Emacs Dev Env", "ts": "TypeScript", "yaml": "YAML", "sml": "Standard ML (SML)", "adb": "Ada", "tsx": "TypeScript", "vbproj": "MSBuild", "pcc": "C++", "dockerfile": "Dockerfile", "yml": "YAML", "ly": "Happy", "coffee": "CoffeeScript", "rx": "Forth", "vert": "GLSL", "psl": "PSL Assertion", "md": "Markdown", "rake": "Rakefile", "ahk": "AutoHotKey", "dockerignore": "Dockerfile", "dart": "Dart", "bat": "Batch", "d": "D", "lisp": "Lisp", "h": "C Header", "cmd": "Batch", "abap": "ABAP", "c++": "C++", "p": "Prolog", "ml": "OCaml", "x": "Alex", "sitemap": "ASP.NET", "el": "Emacs Lisp", "cfm": "ColdFusion", "htm": "HTML", "cfc": "ColdFusion CFScript", "ec": "C", "tex": "TeX", "ex": "Elixir", "btm": "Batch", "ur": "Ur/Web", "targets": "MSBuild", "webinfo": "ASP.NET", "for": "FORTRAN Legacy", "scala": "Scala", "lean": "Lean", "lua": "Lua", "ceylon": "Ceylon", "gvy": "Groovy", "geom": "GLSL", "rb": "Ruby", "props": "MSBuild", "fth": "Forth", "forth": "Forth", "ftn": "FORTRAN Legacy", "asp": "ASP", "fish": "Fish", "irunargs": "Verilog Args File", "comp": "GLSL", "vala": "Vala", "wl": "Wolfram", "asax": "ASP.NET", "jl": "Julia", "erl": "Erlang", "asa": "ASP", "cxx": "C++", "org": "Org", "js": "JavaScript", "swift": "Swift", "bash": "BASH", "c": "C", "lucius": "Lucius", "xrunargs": "Verilog Args File", "upkg": "Unreal Script", "idr": "Idris", "nix": "Nix", "cogent": "Cogent", "cob": "COBOL", "toml": "TOML", "oz": "Oz", "srt": "SRecode Template", "pde": "Processing", "hbs": "Handlebars", "mad": "Madlang", "markdown": "Markdown", "cabal": "Cabal", "pgc": "C", "cc": "C++", "hex": "HEX", "sql": "SQL", "cpp": "C++", "vim": "Vim Script", "ipp": "C++ Header", "xtend": "Xtend", "e4": "Forth", "cs": "C#", "cr": "Crystal", "txt": "Plain Text", "uci": "Unreal Script", "ckt": "Spice Netlist", "java": "Java", "fsproj": "MSBuild", "hrl": "Erlang", "py": "Python", "makefile": "Makefile", "f83": "Forth", "gtpl": "Groovy", "json": "JSON", "master": "ASP.NET", "asm": "Assembly", "f03": "FORTRAN Modern", "frt": "Forth", "rhtml": "Ruby HTML", "f08": "FORTRAN Modern", "pl": "Perl", "pm": "Perl", "mm": "Objective C++", "hx": "Haxe", "ascx": "ASP.NET", "hs": "Haskell", "xaml": "XAML", "jai": "JAI", "pfo": "FORTRAN Legacy", "aspx": "ASP.NET", "hh": "C++ Header", "vue": "Vue", "fsx": "F#", "f95": "FORTRAN Modern", "rst": "ReStructuredText", "fsscript": "F#", "dtsi": "Device Tree", "groovy": "Groovy", "polly": "Polly", "thy": "Isabelle", "f": "FORTRAN Legacy", "julius": "Julius", "ada": "Ada", "mli": "OCaml", "mk": "Makefile", "kts": "Kotlin", "r": "R", "svg": "SVG", "dts": "Device Tree", "vhd": "VHDL", "qcl": "QCL", "zsh": "Zsh", "sty": "TeX", "def": "Module-Definition", "qml": "QML", "vb": "Visual Basic", "v": "Coq", "uc": "Unreal Script", "vg": "Verilog", "vh": "Verilog", "cassius": "Cassius", "pro": "Prolog", "cbl": "COBOL", "ccp": "COBOL", "inl": "C++ Header", "as": "ActionScript", "cobol": "COBOL", "svh": "SystemVerilog", "in": "Autoconf", "purs": "PureScript", "grt": "Groovy", "pas": "Pascal", "cmake": "CMake", "csh": "C Shell", "proto": "Protocol Buffers", "fpm": "Forth", "tese": "GLSL", "nb": "Wolfram", "hpp": "C++ Header", "s": "Assembly", "tesc": "GLSL", "html": "HTML", "pad": "Ada", "fst": "F*", "hxx": "C++ Header", "text": "Plain Text", "css": "CSS", "frag": "GLSL", "fr": "Forth", "fs": "F#", "ft": "Forth", "nim": "Nim", "urs": "Ur/Web", "y": "Happy", "fb": "Forth", "4th": "Forth", "hlean": "Lean", "cshtml": "Razor", "cljs": "ClojureScript", "mak": "Makefile", "php": "PHP", "jsx": "JSX", "agda": "Agda", "lidr": "Idris", "scss": "Sass", "e": "Specman e", "sass": "Sass", "ss": "Scheme", "fsi": "F#", "rs": "Rust", "m": "Objective C", "f90": "FORTRAN Modern", "cpy": "COBOL", "sh": "Shell", "urp": "Ur/Web Project", "exs": "Elixir", "kt": "Kotlin", "sc": "Scala", "f77": "FORTRAN Legacy", "mustache": "Mustache"}

type Language struct {
	Extensions []string `json:"extensions"`
	Language   string   `json:"language"`
}

func loadDatabase() []Language {
	var database []Language
	data, _ := base64.StdEncoding.DecodeString(database_languages)
	_ = json.Unmarshal(data, &database)
	return database
}

/// Get all the files that exist in the directory
func walkDirectory(root string, output *chan *FileJob) {
	gitignore, gitignoreerror := gitignore.NewGitIgnore(filepath.Join(root, ".gitignore"))

	filepath.Walk(root, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		// Need to exclude git, hg, svn etc...
		if strings.HasPrefix(root, ".git/") {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			if gitignoreerror != nil || !gitignore.Match(filepath.Join(root, info.Name()), false) {

				extension := strings.ToLower(path.Ext(info.Name()))

				if !strings.HasPrefix(info.Name(), ".") && strings.Count(info.Name(), ".") == 1 {
					extension = strings.TrimLeft(extension, ".")
				}

				if extension == "" {
					extension = strings.ToLower(info.Name())
				}

				// TODO this should be a hashmap lookup for the speeds
				// exclude := false
				// for _, ex := range Exclusions {
				// 	if strings.HasSuffix(info.Name(), ex) {
				// 		exclude = true
				// 	}
				// }

				language, ok := ExtensionToLanguage[extension]

				if ok {
					*output <- &FileJob{Location: root, Filename: info.Name(), Extension: extension, Language: language}
				}
			}
		}

		return nil
	})

	close(*output)
}

func fileReaderWorker(input *chan *FileJob, output *chan *FileJob) {
	var wg sync.WaitGroup
	for res := range *input {
		wg.Add(1)
		go func(res *FileJob) {
			content, _ := ioutil.ReadFile(res.Location)
			res.Content = content
			*output <- res
			wg.Done()
		}(res)
	}

	go func() {
		wg.Wait()
		close(*output)
	}()
}

func fileProcessorWorker(input *chan *FileJob, output *chan *FileJob) {
	var wg sync.WaitGroup
	for res := range *input {
		// Do some pointless work
		wg.Add(1)
		go func(res *FileJob) {
			res.Lines = int64(bytes.Count(res.Content, []byte("\n")))   // Fastest way to count newlines
			res.Blank = int64(bytes.Count(res.Content, []byte("\n\n"))) // Cheap way to calculate blanks
			// is it? what about the langage "whitespace" where whitespace is significant....

			// Find first instance of a \n
			// Check the slice before for interesting
			// Determine if newline
			// keep running counter
			// check if spaces etc....

			res.Bytes = int64(len(res.Content))
			*output <- res
			wg.Done()
		}(res)
	}

	go func() {
		wg.Wait()
		close(*output)
	}()
}

func fileSummerize(input *chan *FileJob) {

	// Once done lets print it all out
	output := []string{
		"-----",
		"Language | Files | Lines | Code | Comment | Blank | Byte",
		"-----",
	}

	languages := map[string]LanguageSummary{}

	// TODO declare type to avoid cast
	sumFiles := int64(0)
	sumLines := int64(0)
	sumCode := int64(0)
	sumComment := int64(0)
	sumBlank := int64(0)
	sumByte := int64(0)

	for res := range *input {
		sumFiles++
		sumLines += res.Lines
		sumCode += res.Code
		sumComment += res.Comment
		sumBlank += res.Blank
		sumByte += res.Bytes

		// TODO this is SLOW refactor to use pre-generated hashmap lookups
		// its probably not a huge issue because this runs as things come in but
		// lets not make the CPU work harder than it needs to
		_, ok := languages[res.Language]

		if !ok {
			languages[res.Language] = LanguageSummary{
				Name:    res.Language,
				Bytes:   res.Bytes,
				Lines:   res.Lines,
				Code:    res.Code,
				Comment: res.Comment,
				Blank:   res.Blank,
				Count:   1,
			}
		} else {
			tmp := languages[res.Language]

			languages[res.Language] = LanguageSummary{
				Name:    res.Language,
				Bytes:   tmp.Bytes + res.Bytes,
				Lines:   tmp.Lines + res.Lines,
				Code:    tmp.Code + res.Code,
				Comment: tmp.Comment + res.Comment,
				Blank:   tmp.Blank + res.Blank,
				Count:   tmp.Count + 1,
			}
		}
	}

	for name, summary := range languages {
		output = append(output, fmt.Sprintf("%s | %d | %d | %d | %d | %d | %d", name, summary.Count, summary.Lines, summary.Code, summary.Comment, summary.Blank, summary.Bytes))
	}

	output = append(output, "-----")
	output = append(output, fmt.Sprintf("Total | %d | %d | %d | %d | %d | %d", sumFiles, sumLines, sumCode, sumComment, sumBlank, sumByte))
	output = append(output, "-----")

	result := columnize.SimpleFormat(output)
	fmt.Println(result)
}

//go:generate go run scripts/include.go
func main() {
	// A buffered channel that we can send work requests on.
	fileReadJobQueue := make(chan *FileJob, runtime.NumCPU()*20)
	fileProcessJobQueue := make(chan *FileJob, runtime.NumCPU())
	fileSummaryJobQueue := make(chan *FileJob, runtime.NumCPU()*20)

	go walkDirectory(os.Args[1], &fileReadJobQueue)
	go fileReaderWorker(&fileReadJobQueue, &fileProcessJobQueue)
	go fileProcessorWorker(&fileProcessJobQueue, &fileSummaryJobQueue)
	fileSummerize(&fileSummaryJobQueue) // Bring it all back to you
}
