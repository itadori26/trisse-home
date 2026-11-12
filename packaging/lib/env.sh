# Shared by the trisse-home helper scripts (sourced, not executed).
#
# trisse_home_value VAR prints the value the service sees for VAR, following
# the order of EnvironmentFile= in the unit: trisse-home.env (admin) wins over
# appliance.env (Raspberry Pi image defaults), which wins over secret.env.

TRISSE_HOME_CONF_DIR=${TRISSE_HOME_CONF_DIR:-/etc/trisse-home}

trisse_home_value() {
	for f in trisse-home.env appliance.env secret.env; do
		v=$(sed -n "s/^[[:space:]]*$1=\([^[:space:]]*\).*/\1/p" "$TRISSE_HOME_CONF_DIR/$f" 2>/dev/null | tail -n 1)
		if [ -n "$v" ]; then
			echo "$v"
			return
		fi
	done
}

# trisse_home_url_suffix prints ":<port>" unless the service listens on 80.
trisse_home_url_suffix() {
	listen=$(trisse_home_value LISTEN_ADDR)
	port=${listen##*:}
	case "$port" in
	"" | 80) ;;
	*) echo ":$port" ;;
	esac
}

# trisse_home_setup_status prints pending, done or unknown (service not up).
trisse_home_setup_status() {
	listen=$(trisse_home_value LISTEN_ADDR)
	env LISTEN_ADDR="${listen:-:8090}" trisse-home setup-status 2>/dev/null || echo unknown
}
