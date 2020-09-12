#!/usr/bin/env bash
echo "$settings" > /config.toml
echo "setting:" $settings

/api-service -c config.toml
