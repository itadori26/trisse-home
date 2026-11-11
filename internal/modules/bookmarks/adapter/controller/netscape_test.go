package controller

import (
	"strings"
	"testing"
)

const firefoxExport = `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks Menu</H1>
<DL><p>
    <DT><A HREF="https://go.dev/" ADD_DATE="1700000000" TAGS="golang,docs">Go &amp; friends</A>
    <DD>The Go website
    <DT><H3 ADD_DATE="1" PERSONAL_TOOLBAR_FOLDER="true">Toolbar</H3>
    <DL><p>
        <DT><H3>Work</H3>
        <DL><p>
            <DT><A HREF="https://example.com/wiki">Wiki</A>
        </DL><p>
        <DT><A HREF="javascript:alert(1)">Bookmarklet</A>
    </DL><p>
    <DT><A HREF="https://news.example/">News</A>
</DL><p>`

func TestParseNetscape(t *testing.T) {
	items, err := ParseNetscape(strings.NewReader(firefoxExport))
	if err != nil || len(items) != 4 {
		t.Fatalf("%+v %v", items, err)
	}
	if items[0].Title != "Go & friends" || items[0].Description != "The Go website" || len(items[0].Tags) != 2 || items[0].FolderPath != "" {
		t.Fatalf("first: %+v", items[0])
	}
	if items[1].FolderPath != "Toolbar / Work" || items[2].FolderPath != "Toolbar" || items[3].FolderPath != "" {
		t.Fatalf("folders: %q %q %q", items[1].FolderPath, items[2].FolderPath, items[3].FolderPath)
	}
}
