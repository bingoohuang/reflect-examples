# Sudo

sudo is a package to make reflect more powerful (and dangerous), original from [here](github.com/zeebo/sudo).

It exports a single function,  `Sudo`, which when passed a `reflect.Value`, will return a new `reflect.Value` with the read-only restrictions removed.
