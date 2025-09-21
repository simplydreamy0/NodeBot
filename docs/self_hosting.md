# Self-hosting

This project is open source, you can create your own bot user account and self-host this bot using your bot user account if you want. The CI/CD of this project already builds a docker image for differents architectures you just need to deploy it somewhere of you choice.

## Generating bot tokens

Because of how Twitch API authorization is done, our bot needs multiple types of tokens: app and user tokens.

### App acces token

After configuring twitch CLI, launch:
```
twitch token
```

### User access token

After configuring twitch CLI, launch:
```
twitch token --user-token --scopes "user:read:chat user:write:chat user:bot moderator:manage:shoutouts moderator:manage:announcements"
```
When the window shows up, connect with the bot user account & authorize the app

## Handling TLS

This bot uses [Twitch's event sub](https://dev.twitch.tv/docs/eventsub/) and Webhooks as transport, this means your bot needs to be hosted publicly behind a TLS endpoint for twitch to send you the events the bot subscribes to.

There are a multitude of ways to deploy a webhook behind a TLS endpoint, here's a couple:
* Buy a compute instance & self host with docker + cert-manager + let's encrypt
* Vercel, fly.io, Heroku & other serverless platforms
