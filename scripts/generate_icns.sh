#!/usr/bin/env bash
#
# generate_icns.sh -- Generate necessary images that can be compiled into a macOS `icns` icon file
#
# constants
declare -a DIMENSIONS=(16 32 32 64 128 256 256 512 512 1024)
declare -a DIMENSION_LABEL=(
    "16x16"
    "16x16@2x"
    "32x32"
    "32x32@2x"
    "128x128"
    "128x128@2x"
    "256x256"
    "256x256@2x"
    "512x512"
    "512x512@2x"
)
#
# functions
usage() {
    printf "usage:\n"
    printf "\t%s <SVG file> <output path>\n" "$0"
}
#
# entry point
if [[ -z "$1" ]]; then
    usage
    exit 1
fi
if [[ ! -f "$1" ]]; then
    printf "*  error: input file '%s' does not exist.\n" "$1"
    exit 2
fi
INPUT_FILE="$1"
# take the base name of the input file
BASE_NAME=$(basename "$INPUT_FILE")
# lower-case the base name
BASE_NAME=${BASE_NAME,,}
# remove any trailing SVG extension
BASE_NAME=${BASE_NAME%%.svg}
if [[ -z "$2" ]]; then
    usage
    exit 1
fi
OUTPUT_PATH="$2"
# look for inkscape
INKSCAPE_CMD="inkscape"
if ! hash "$INKSCAPE_CMD" &>/dev/null; then
    if [[ -d "/Applications/Inkscape.app/Contents/MacOS" ]]; then
        INKSCAPE_CMD="/Applications/Inkscape.app/Contents/MacOS/inkscape"
    else
        printf "*  error: cannot find inkscape; inkscape is required for image conversion"
        exit 3
    fi
fi
# create output path
if ! mkdir -p "$OUTPUT_PATH"; then
    printf "*  error: cannot create output directory '%s'.\n" "$OUTPUT_PATH"
    exit 3
fi
# create iconset directory
ICONSET_PATH="$OUTPUT_PATH/$BASE_NAME.iconset"
if ! mkdir "$ICONSET_PATH"; then
    printf "*  error: cannot create iconset directory '%s'.\n" "$ICONSET_PATH"
    exit 4
fi
# loop through dimensions to create images
for DIM_INDEX in "${!DIMENSIONS[@]}"; do
    DIMENSION=${DIMENSIONS[$DIM_INDEX]}
    ICON_FILENAME="icon_${DIMENSION_LABEL[$DIM_INDEX]}.png"
    ICON_PATH="$ICONSET_PATH/$ICON_FILENAME"
    printf "creating %s\n" "$ICON_FILENAME"
    $INKSCAPE_CMD -w "$DIMENSION" -h "$DIMENSION" "$INPUT_FILE" -o "$ICON_PATH" || true
done
# create ICNS file
ICNS_PATH="$OUTPUT_PATH/$BASE_NAME.icns"
printf "creating %s\n" "$BASE_NAME.icns"
iconutil -c icns -o "$ICNS_PATH" "$ICONSET_PATH"
# all done.
printf "done.\n"
