#!/bin/sh

PEER="31.134.138.189:56000"
VK_LINK="https://vk.com/call/join/KxFnVKH3iUy0AT8R9lyC3QTEmA2mnYDr52Fo6r9WFKE"
LISTEN="127.0.0.1:9000"
LOG="client.log"
BIN="client-bin"

echo "Starting vk-turn-proxy..."
echo "Peer:   $PEER"
echo "Listen: $LISTEN"
echo "Log:    $LOG"
echo ""

exec "$BIN" \
  -peer "$PEER" \
  -vk-link "$VK_LINK" \
  -listen "$LISTEN" \
  2>&1 | tee "$LOG"
