# TrueNAS Tailscale Companion

`truenas-tailscale` is a companion program for TrueNAS installs.

It will make your TrueNAS install and app portals, available on your Tailnet automagically, complete with LetsEncrypt certificates.

## Installation

1. Configure a TrueNAS key. In the TrueNAS UI: Settings > API Keys > Add.
 - Settings is in the top-right of the TrueNAS UI, under the user profile pic, next to the power icon.
 - You can also find it here: https://truenas/ui/apikeys (use your own truenas hostname).
2. Configure a TailScale auth key in the [web portal](https://login.tailscale.com/admin/settings/keys) under Settings > Keys.
 - Reusable: True
 - Expiration: 90 days (Note: you will need a new key every 90 days)
 - Ephemeral: True (Recommended)
 - Tags: (optional - for access control)
3. Create a file called `tailscale-start.sh` that looks like this:
```
export TS_AUTHKEY=tskey-auth-abc123...
export TS_HOSTNAME=truenas-tailscale
export TRUENAS_API_KEY=1-OKabc123...

pgrep truenas-tailscale && exit 1

nohup /mnt/tank-1/path/to/truenas-tailscale >> /mnt/tank-1/path/to/truenas-tailscale.log &
```
4. Copy the `tailscale-start.sh` script and the [binary](github.com/dwurf/truenas-tailscale/releases/latest) onto your NAS. Set the execute bit on both files.
5. Under System > Advanced Settings > Init/Shutdown Scripts, create a new script:
 - Description: `truenas-tailscale`
 - Type: `Script`
 - Script: `/mnt/tank-1/path/to/tailscale-start.sh`
 - When: `Post Init`
 - Enabled: `True`
 - Timeout: `10`

## Detailed Usage

```
$ truenas-tailscale -h
Usage of truenas-tailscale:
  -tailscale-api-key string
    	Tailscale API Key (env: TS_AUTHKEY).
  -tailscale-hostname string
    	Hostname to use in the tailnet. Defaults to the hostname configured in TrueNAS (env: TS_HOSTNAME).
  -truenas-api-key string
    	TrueNAS API key (env: TRUENAS_API_KEY).
  -truenas-hostname string
    	TrueNAS hostname or IP (env: TRUENAS_HOSTNAME). (default "127.0.0.1")
```

Example:
```
$ export TS_AUTHKEY=tskey-auth-abc123 TRUENAS_API_KEY=1-OKabc123
$ truenas-tailscale -tailscale-hostname my-truenas-host
```

