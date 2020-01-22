# crumb

<a href="">
  <img src="https://github.com/svaante/crumb/blob/master/.github/ludwig_richter.jpg" align="right" />
</a>

### What is crumb?

`crumb` is a cli tool adding and managing crumbs your `WD`. So what is crumbs you might ask? Its whatever you want it to be. 

A crumb inside a crumb file follows has the following syntax.. no more no less
`Creation Date` `[Modify Date]` `[Marking]` `Text`

So in essence its a `Text` string with some metadata or the date stuff is mostly accidental, so only a `Marking` tag. So its up to your understanding of the `crumbs` to `Marking`.

Lets define our crumbs as todo items:
```sh
$ cat > $HOME/.crumb.toml
[Markers.todo]
Prefix='"[ ] "'

[Markers.done]
Prefix='"[x] "'

[Markers.selected]
Prefix='" >  "'

[Markers.backlog]
Prefix='" ~  "'
Suffix='"  ~"'
^D
$ cd /home/svaante/project1
$ crumb ad todo This is what i need to do
$ crumb ls
/home/svaante/project1
[ ] This is what i need to do
$ cd ..
$ crumb ad done Started todo list for project1
$ crumb fl
/home/svaante
[x] Started todo list for project1
/home/svaante/project1
[ ] This is what i need to do
$ crumb ma project1 selected 1
$ crumb fl
/home/svaante
[x] Started todo list for project1
/home/svaante/project1
 >  This is what i need to do
```

## 
