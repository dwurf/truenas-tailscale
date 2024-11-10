# TrueNAS Tailscale Companion

`truenas-tailscale` is a companion program for TrueNAS installs.

It will make your TrueNAS install and app portals available on your Tailnet automagically, complete with LetsEncrypt certificates.

## Why?

Compared to installing the Tailscale app, this tool gives you:

- Internal DNS names for each of your apps.
- Automatic TLS for your apps and TrueNAS UI.

## Limitations

- Only forwards the first Portal for a given App.
- No support for other services (such as SSH, MinIO) over the Tailnet right now.
- Tailscale Auth Key must be renewed every 90 days right now.

## Installation

0. Sign up for Tailscale (if you haven't already). Enable [MagicDNS and HTTPS](https://tailscale.com/kb/1153/enabling-https).
1. Configure a Tailscale auth key in the [web portal](https://login.tailscale.com/admin/settings/keys) under Settings > Keys.
 - Reusable: `True`.
 - Expiration: `90 days` (Note: you will need a new key every 90 days).
 - Ephemeral: `True` (Recommended).
 - Tags: `tag:truenas-tailscale` (optional - for access control and management).
2. Configure a TrueNAS API key. In the TrueNAS UI: Settings > API Keys > Add.
 - Settings is in the top-right of the TrueNAS UI, under the user profile pic, next to the power icon.
 - You can also find it here: https://truenas/ui/apikeys (use your own truenas hostname).
3. Create a TrueNAS custom app under Apps > Discover Apps > Custom App.
 - Application Name: `truenas-tailscale`.
 - Repository: `dwurf/truenas-tailscale`.
 - Environment Variables
   - `TS_AUTHKEY`: set to the key you created in step 1.
   - `TS_HOSTNAME`: `truenas-tailscale` (or whatever you want your node name to be called in Tailscale).
   - `TRUENAS_API_KEY`: set to the key you created in step 2.
   - `TRUENAS_HOSTNAME`: set to `172.16.0.1` (the default IP for truenas from within Docker).
 - Restart Policy: `Unless Stopped`.
 - Create a Storage Configuration.
   - Type `ixVolume`.
   - Read Only: `False`.
   - Mount Path: `/root/.config/truenas-tailscale`.
   - Dataset Name: `tailscale-config`.

## Detailed Usage

The command can be downloaded and run directly from the Releases page, see instructions below.

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

Note: the command can be run anywhere that can connect to your NAS, it doesn't have to be directly on the NAS.
