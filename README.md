# capture-all-the-scripts

This thing started as an experiment to see how many ssh bruteforceing bots i 
could keep hanging around, at first my thought was to send the ssh banner
really slowly to see if these scripts do any timeout

But as bandwidth is cheap it ended up trying to send as much ssh banner as 
quickly as possible to the bots wasting both their time and bandwidth - and also my time and bandwidth :)

This is what it looks like


This thing doesn't have much code quality as its just slapped together, if you would like to run it your self, it goes something like this:

```

# move your existing ssh daemon to another port
go get github.com/fasmide/capture-all-the-scripts
# create an ssh key make sure to save it in the same directory as this
ssh-keygen
# have some large text document, save it as ebook.txt
# you are now ready to go
./capture-all-the-scripts -port 22
# if your planning on running this thing for long, use something like tmux

```
