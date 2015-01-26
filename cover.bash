#!/bin/bash -e

GO=${GO:-go}
MODULES=( "$@" )

if [[ ${#MODULES[@]} == 0 ]]; then
    echo "Usage:"
    echo "$0 <modules>"
    exit 1
fi

DIR="`pwd`/cover_report"
[ -d "$DIR" ] && rm -rf "$DIR"
mkdir "$DIR"

cat << EOF > "$DIR/index.html"
<!DOCTYPE html>
<html>
<head>
</head>
<body>
<ul>
EOF

for (( i=0; i<${#MODULES[@]}; i++ ));
do
    echo "=== ${MODULES[$i]}"
    module=`echo -n ${MODULES[$i]} | sed 's#/#_#g'`
    $GO test -cover -coverprofile "$DIR/$module.out" "${MODULES[$i]}"
    $GO tool cover -html="$DIR/$module.out" -o "$DIR/$module.html"
    echo "<li><a href=\"$DIR/$module.html\">$module</a></li>" >> "$DIR/index.html"
done

cat << EOF >> "$DIR/index.html"
</ul>
</body>
</html>
EOF

echo "Coverage report: $DIR/index.html"
