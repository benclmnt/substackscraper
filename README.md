# Substack Scraper

A small cli tool to grab substack posts (free or paid) and outputs it into markdown files.

```
Usage of substackscraper:
  -cookie substack.sid
        Substack API cookie (remove the substack.sid prefix)
  -dest string
        Destination folder to write output to (default ".")
  -output string
        Output format: html or md (default "md")
  -pub string
        Name of the Substack publication to scrape (required)
  -since string
        Fetch posts since this date (YYYY-MM-DD) (default "1970-01-01")
```

## Usage

Building:

```
git clone https://github.com/benclmnt/substackscraper
go build
```

Running:

You will need to grab your substack cookie from your browser. Without the cookie, it can only fetch the public preview of paid posts.

```
./substackscraper -pub benclmnt -cookie COOKIE_SUBSTACK.SID_VALUE_ONLY
```

## Note

This CLI relies on a few undocumented substack API version 1. The endpoints might change anytime, breaking the tool.