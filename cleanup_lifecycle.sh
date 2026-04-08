#!/bin/bash
# Cleanup all lifecycle-test objects — uses grep instead of python3 for ref extraction
set -o pipefail

WAPI="https://10.34.33.115/wapi/v2.12.3"
AUTH="admin:infoblox"

delete_obj() {
  local query="$1"
  local label="$2"
  local resp
  resp=$(curl -k -s -u "$AUTH" "$WAPI/$query")
  # Extract _ref using grep/sed (no python3 needed)
  local ref
  ref=$(echo "$resp" | grep -o '"_ref": *"[^"]*"' | head -1 | sed 's/"_ref": *"//;s/"$//')
  if [ -n "$ref" ] && [ "$ref" != "[]" ]; then
    local del_resp
    del_resp=$(curl -k -s -u "$AUTH" -X DELETE "$WAPI/$ref")
    echo "[OK] Deleted $label -> $del_resp"
  else
    echo "[--] Not found: $label"
  fi
}

echo "=== 1. Smart Folders ==="
delete_obj "smartfolder:global?name=lifecycle-test-smartfolder" "smartfolder:global"

echo ""
echo "=== 2. VLANs ==="
delete_obj "vlan?name=lifecycle-test-vlan" "vlan"
delete_obj "vlanrange?name=lifecycle-test-vlanrange" "vlanrange"
delete_obj "vlanview?name=lifecycle-test-vlanview" "vlanview"

echo ""
echo "=== 3. DTC (lbdn -> pool -> server/monitors) ==="
delete_obj "dtc:lbdn?name=lbdn-lifecycle.lifecycle-test.example.com" "dtc:lbdn"
delete_obj "dtc:pool?name=lifecycle-test-dtc-pool" "dtc:pool"
delete_obj "dtc:server?name=lifecycle-test-dtc-server" "dtc:server"
delete_obj "dtc:monitor:icmp?name=lifecycle-test-dtc-icmp" "dtc:monitor:icmp"
delete_obj "dtc:monitor:http?name=lifecycle-test-dtc-http" "dtc:monitor:http"
delete_obj "dtc:monitor:tcp?name=lifecycle-test-dtc-tcp" "dtc:monitor:tcp"
delete_obj "dtc:topology?name=lifecycle-test-dtc-topo" "dtc:topology"

echo ""
echo "=== 4. Shared Records ==="
delete_obj "sharedrecord:a?name=shared-a.lifecycle-test.example.com" "sharedrecord:a"
delete_obj "sharedrecord:txt?name=shared-txt.lifecycle-test.example.com" "sharedrecord:txt"
delete_obj "sharedrecordgroup?name=lifecycle-test-srg" "sharedrecordgroup"

echo ""
echo "=== 5. ACL ==="
delete_obj "namedacl?name=lifecycle-test-acl" "namedacl"

echo ""
echo "=== 6. Security & Grid ==="
delete_obj "admingroup?name=lifecycle-test-group" "admingroup"
delete_obj "adminrole?name=lifecycle-test-role" "adminrole"
delete_obj "extensibleattributedef?name=lifecycle-test-ea" "extensibleattributedef"
delete_obj "natgroup?name=lifecycle-test-natgroup" "natgroup"

echo ""
echo "=== 7. DHCP ==="
delete_obj "fixedaddress?ipv4addr=10.200.0.10&network_view=lifecycle-test-nv" "fixedaddress"
delete_obj "range?start_addr=10.200.1.100&network_view=lifecycle-test-nv" "range"
delete_obj "dhcpoptiondefinition?name=lifecycle-test-option" "dhcpoptiondefinition"
delete_obj "dhcpoptionspace?name=lifecycle-test-space" "dhcpoptionspace"
delete_obj "filtermac?name=lifecycle-test-mac-filter" "filtermac"

echo ""
echo "=== 8. DNS Records ==="
for REC in \
  "record:a?name=a-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:a?name=a-ttl.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:aaaa?name=aaaa-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:cname?name=cname-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:mx?name=lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:txt?name=txt-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:txt?name=txt-update.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:srv?name=_sip._tcp.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:naptr?name=lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:caa?name=lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:alias?name=alias-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:dname?name=dname-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view" \
  "record:tlsa?name=_443._tcp.a-lifecycle.lifecycle-test.example.com&view=lifecycle-test-view"
do
  TYPE=$(echo "$REC" | cut -d'?' -f1)
  delete_obj "$REC" "$TYPE"
done

echo ""
echo "=== 9. DNS Zones ==="
delete_obj "zone_delegated?fqdn=delegated.lifecycle-test.example.com&view=lifecycle-test-view" "zone_delegated"
delete_obj "zone_forward?fqdn=forward.lifecycle-test.example.com&view=lifecycle-test-view" "zone_forward"
delete_obj "zone_auth?fqdn=lifecycle-test.example.com&view=lifecycle-test-view" "zone_auth"

echo ""
echo "=== 10. IPAM ==="
delete_obj "ipv6networkcontainer?network=2001%3Adb8%3A200%3A%3A%2F40&network_view=lifecycle-test-nv" "ipv6networkcontainer"
delete_obj "ipv6network?network=2001%3Adb8%3A100%3A%3A%2F48&network_view=lifecycle-test-nv" "ipv6network"
delete_obj "networkcontainer?network=10.201.0.0%2F16&network_view=lifecycle-test-nv" "networkcontainer"
delete_obj "network?network=10.200.1.0%2F24&network_view=lifecycle-test-nv" "network 10.200.1.0/24"
delete_obj "network?network=10.200.0.0%2F24&network_view=lifecycle-test-nv" "network 10.200.0.0/24"

echo ""
echo "=== 11. Views (last) ==="
delete_obj "networkview?name=lifecycle-test-nv" "networkview"
delete_obj "view?name=lifecycle-test-view" "dns view"

echo ""
echo "=== Cleanup complete ==="
