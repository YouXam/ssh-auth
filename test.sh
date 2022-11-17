./ssh-auth server add -i ~/.ssh/id_rsa -n AAA youxam@10.122.208.40
./ssh-auth server add -n BBB test@10.122.208.40
./ssh-auth user add youxam ~/.ssh/id_ed25519_clear.pub
./ssh-auth auth add AAA youxam
./ssh-auth auth add BBB youxam