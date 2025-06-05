# Since we are in the asset folder, move two
# folders up to get the program name

dir=$(pwd)/..
program=$(basename $(builtin cd $dir; pwd))

# The folders we are working with

codeFolder="/home/per/code"
binFolder="/home/per/bin"
softtubeFolder="/softtube/bin"
configFolder="/home/per/.config/softteam"
# cacheFolder="/home/per/.cache/"
desktopFolder="/home/per/.local/share/applications"

# Perform the copy

cp -rf $codeFolder/$program/build/* $binFolder/$program
cp -rf $codeFolder/$program/build/* $softtubeFolder/$program
# cp -rf $codeFolder/$program/assets $binFolder/$program
# cp -rf $codeFolder/$program/assets $softtubeFolder/$program
cp -rf $configFolder/softimdb $softtubeFolder/$program/config
# cp -rf $cacheFolder/softimdb $softtubeFolder/$program/cache
cp -rf $desktopFolder/softimdb.desktop $softtubeFolder/$program
