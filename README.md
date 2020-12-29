## Post-Digress
Terminal UI application for PostgreSQL.

![Demo](https://github.com/Felipe-Aquino/postdigress/blob/master/demonstration.png)


Written in Go, and forged from my struggles and frustations with psql, pgadmin and PostBird.
It's intended to be simple, vi-based, faster than the overbloated ones and allow you to make you queries with ease.

### Fetch'em
This project would'nt be possible without the following Go packages: [pq](https://github.com/lib/pq), [tview](https://github.com/rivo/tview), [tcell](https://github.com/gdamore/tcell).

```bash
go get "github.com/lib/pq"
go get "github.com/rivo/tview"
go get "github.com/gdamore/tcell"
```

### Mind you
Mouse is disabled for the aplication, at least for this release.

Once you estabilish a connection with the database, the main page will be open to you.
Notice that the menu has the focus and will be receiving any key event,
and will make transition between the pages **Execute**, **Structure**,
**Help** and **Quit**, each one indicating(underline) the key that should be pressed to
make the transition to that page.

### Pages
* Execute: Has an editor, a table viewer and a status bar at bottom.
  1. Press Ctrl-E to enter the editor. You can navigate thought the text using
  the following vi-like keybindings _h_, _j_, _k_, _l_, _w_, _e_, _b_, _i_, _a_,
  more yet to come. Press _v_ and _j_ or _k_ to select the queries you wish to execute.
  Press _q_, on select mode, to put you back on normal mode.
  Press _q_, on normal mode, to exit the editor, this will put the menu on focus.

  2. Press Ctrl-T, to put focus on the table. Use _m_ to change the navigation mode,
  that can be cell, row or column. You can use vi-like keybindings _h_, _j_, _k_, _l_ to navigate
  the table. Press _q_, to exit the table, this will put the menu on focus.

  3. The status bar should containt useful informations about the editor and/or the table

* Structure: Has 3 panes, to show the tables and the columns and constraints of a selected table.
  1. Press _d_ following of _j_, _k_ to navigate the database tables. Hit enter to select one of 
  the tables, informations about that table will be queried and should be visible in the other 2 panes.

  2. Press _c_ or _i_ to focus on the other panes. You can use the vi-like keys to navigate.

* Help: Show this text that you area currently reading
* Quit: Quits the application

### Tricks
In the connection page you can use Tab, Ctrl-J, Ctrl-K, Ctrl-L, Ctrl-H to move between the form fields

In the execute page, with the menu on focus you can press _0_ or _9_ to set the focus on the editor or the table.
These keys and _q_ can speed your life in this application.
