#!/bin/bash

# Test script to validate command alias alignment
# This ensures all aliases support the same flags as their primary commands

set -e

echo "🔍 Testing Command Alias Alignment..."
echo

# Build the tool first
go build -o vultool ./cmd/vultool

# Test files
UNENCRYPTED_FILE="test/fixtures/testGG20-part1of2.vult"
ENCRYPTED_FILE="test/fixtures/qa-fast-share2of2.vult"
ENCRYPTED_PASSWORD="vulticli01"
VAULT1="test/fixtures/testGG20-part1of2.vult"
VAULT2="test/fixtures/testGG20-part2of2.vult"

echo "✅ Testing 'info' alias (should work like 'inspect --summary'):"
echo "Running: ./vultool info -f $UNENCRYPTED_FILE"
./vultool info -f "$UNENCRYPTED_FILE" > /tmp/info_output.txt
echo "Running: ./vultool inspect -f $UNENCRYPTED_FILE --summary"
./vultool inspect -f "$UNENCRYPTED_FILE" --summary > /tmp/inspect_summary_output.txt

if diff /tmp/info_output.txt /tmp/inspect_summary_output.txt > /dev/null; then
    echo "✅ INFO ALIAS: Outputs match ✓"
else
    echo "❌ INFO ALIAS: Outputs differ ✗"
    echo "Differences:"
    diff /tmp/info_output.txt /tmp/inspect_summary_output.txt || true
fi

echo
echo "✅ Testing 'info' alias with password support:"
echo "Running: ./vultool info -f $ENCRYPTED_FILE --password $ENCRYPTED_PASSWORD"
./vultool info -f "$ENCRYPTED_FILE" --password "$ENCRYPTED_PASSWORD" > /tmp/info_encrypted_output.txt
if [[ -s /tmp/info_encrypted_output.txt ]] && ! grep -q "Error" /tmp/info_encrypted_output.txt; then
    echo "✅ INFO ALIAS: Password support works ✓"
else
    echo "❌ INFO ALIAS: Password support failed ✗"
fi

echo
echo "✅ Testing 'verify' alias (should work like 'inspect --validate'):"
echo "Running: ./vultool verify -f $UNENCRYPTED_FILE"
./vultool verify -f "$UNENCRYPTED_FILE" > /tmp/verify_output.txt 2>&1
echo "Running: ./vultool inspect -f $UNENCRYPTED_FILE --validate"
./vultool inspect -f "$UNENCRYPTED_FILE" --validate > /tmp/inspect_validate_output.txt 2>&1

if diff /tmp/verify_output.txt /tmp/inspect_validate_output.txt > /dev/null; then
    echo "✅ VERIFY ALIAS: Outputs match ✓"
else
    echo "❌ VERIFY ALIAS: Outputs differ ✗"
    echo "Differences:"
    diff /tmp/verify_output.txt /tmp/inspect_validate_output.txt || true
fi

echo
echo "✅ Testing 'verify' alias with password support:"
echo "Running: ./vultool verify -f $ENCRYPTED_FILE --password $ENCRYPTED_PASSWORD"
./vultool verify -f "$ENCRYPTED_FILE" --password "$ENCRYPTED_PASSWORD" > /tmp/verify_encrypted_output.txt 2>&1
if [[ -s /tmp/verify_encrypted_output.txt ]] && ! grep -q "Error" /tmp/verify_encrypted_output.txt; then
    echo "✅ VERIFY ALIAS: Password support works ✓"
else
    echo "❌ VERIFY ALIAS: Password support failed ✗"
fi

echo
echo "✅ Testing 'decode' alias (JSON output):"
echo "Running: ./vultool decode -f $UNENCRYPTED_FILE"
./vultool decode -f "$UNENCRYPTED_FILE" > /tmp/decode_json_output.txt
if [[ -s /tmp/decode_json_output.txt ]] && jq . /tmp/decode_json_output.txt > /dev/null 2>&1; then
    echo "✅ DECODE ALIAS: JSON output works ✓"
else
    echo "❌ DECODE ALIAS: JSON output failed ✗"
fi

echo
echo "✅ Testing 'decode' alias (YAML output):"
echo "Running: ./vultool decode -f $UNENCRYPTED_FILE --yaml"
./vultool decode -f "$UNENCRYPTED_FILE" --yaml > /tmp/decode_yaml_output.txt
if [[ -s /tmp/decode_yaml_output.txt ]] && grep -q "name:" /tmp/decode_yaml_output.txt; then
    echo "✅ DECODE ALIAS: YAML output works ✓"
else
    echo "❌ DECODE ALIAS: YAML output failed ✗"
fi

echo
echo "✅ Testing 'decode' alias with password support:"
echo "Running: ./vultool decode -f $ENCRYPTED_FILE --password $ENCRYPTED_PASSWORD"
./vultool decode -f "$ENCRYPTED_FILE" --password "$ENCRYPTED_PASSWORD" > /tmp/decode_encrypted_output.txt
if [[ -s /tmp/decode_encrypted_output.txt ]] && jq . /tmp/decode_encrypted_output.txt > /dev/null 2>&1; then
    echo "✅ DECODE ALIAS: Password support works ✓"
else
    echo "❌ DECODE ALIAS: Password support failed ✗"
fi

echo
echo "✅ Testing 'diff' command:"
echo "Running: ./vultool diff $VAULT1 $VAULT2"
./vultool diff "$VAULT1" "$VAULT2" > /tmp/diff_output.txt
if [[ -s /tmp/diff_output.txt ]] && (grep -q "differ" /tmp/diff_output.txt || grep -q "identical" /tmp/diff_output.txt); then
    echo "✅ DIFF COMMAND: Basic functionality works ✓"
else
    echo "❌ DIFF COMMAND: Basic functionality failed ✗"
fi

echo
echo "✅ Testing 'diff' command with JSON output:"
echo "Running: ./vultool diff --json $VAULT1 $VAULT2"
./vultool diff --json "$VAULT1" "$VAULT2" > /tmp/diff_json_output.txt
if [[ -s /tmp/diff_json_output.txt ]] && jq . /tmp/diff_json_output.txt > /dev/null 2>&1; then
    echo "✅ DIFF COMMAND: JSON output works ✓"
else
    echo "❌ DIFF COMMAND: JSON output failed ✗"
fi

echo
echo "✅ Testing 'diff' command with YAML output:"
echo "Running: ./vultool diff --yaml $VAULT1 $VAULT2"
./vultool diff --yaml "$VAULT1" "$VAULT2" > /tmp/diff_yaml_output.txt
if [[ -s /tmp/diff_yaml_output.txt ]] && grep -q ":" /tmp/diff_yaml_output.txt; then
    echo "✅ DIFF COMMAND: YAML output works ✓"
else
    echo "❌ DIFF COMMAND: YAML output failed ✗"
fi

echo
echo "✅ Testing 'diff' command with password support:"
echo "Running: ./vultool diff --password $ENCRYPTED_PASSWORD test/fixtures/qa-fast-share1of2.vult $ENCRYPTED_FILE"
./vultool diff --password "$ENCRYPTED_PASSWORD" test/fixtures/qa-fast-share1of2.vult "$ENCRYPTED_FILE" > /tmp/diff_encrypted_output.txt
if [[ -s /tmp/diff_encrypted_output.txt ]] && (grep -q "differ" /tmp/diff_encrypted_output.txt || grep -q "identical" /tmp/diff_encrypted_output.txt); then
    echo "✅ DIFF COMMAND: Password support works ✓"
else
    echo "❌ DIFF COMMAND: Password support failed ✗"
fi

echo
echo "🎯 Summary: Command alias alignment validation complete!"
echo

# Cleanup
rm -f /tmp/*_output.txt

echo "All alias tests completed. Check the output above for any issues."
