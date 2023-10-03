#!/usr/bin/env bash

#
# This script helps how to generate many reports
# to analyse the data together.
# The outcome is the extracted and processed OPCT data
# and a index file to be explored.
# Run the file server to explore it: $ python3 -m http.server 3333
#

export RESULTS=();
RESULTS+=( ocp414rc0_AWS_None_202309222127 )
RESULTS+=( ocp414rc2_AWS_AWS_202309290029 )
RESULTS+=( ocp414rc2_AWS_None_202309282349 )
RESULTS+=( ocp414rc2_Azure_Azure-IPI_202310022324  )
RESULTS+=( ocp414rc2_Azure_Azure-IPI_202309291531 )
RESULTS+=( ocp414rc2_Azure_Azure-IPI_202309290322 )
RESULTS+=( ocp414rc2_Azure_Azure-etcd-dedicated_202310010646 )
RESULTS+=( ocp414rc2_Azure_Azure-etcd-dedicated_202310020612 )
RESULTS+=( ocp414rc2_Azure_Azure-etcd-dedicated_202310021500 )
RESULTS+=( ocp414rc2_OCI-120vpu1200gib_External_202309281802 )
RESULTS+=( ocp414rc0_OCI_External_202309221951 )
RESULTS+=( ocp414rc0_OCI_None_202309230246 )

result_path=/tmp/results-shared/REPORT_AZURE
mkdir -pv "$result_path"
cat <<EOF > "$result_path"/index.html
<!DOCTYPE html>
<html lang="en">
EOF

for RES in ${RESULTS[*]}; do 
  echo "CREATING $RES";
  echo "<p><a href=\"./${RES}/opct-report.html\">${RES}</a></p>" >> "$result_path"/index.html
  #mkdir -pv "$result_path"/$RES ; ~/opct/bin/opct-devel report --server-skip --save-to "$result_path"/$RES $RES;
done

cat <<EOF >> "$result_path"/index.html
</body>
</html>
EOF

# https://openshift-provider-certification.s3.us-west-2.amazonaws.com/index.html
bucket_name=openshift-provider-certification
bucket_obj_prefix=tmp/report-az
files=( opct-report.html )
files+=( opct-report.json )
files+=( opct-filter.html )
files+=( metrics.html )
files+=( artifacts_must-gather_camgi.html )
files+=( must-gather/event-filter.html )

S3_PREFIXES=s3://$bucket_name/$bucket_obj_prefix
aws s3 cp ${result_path}/index.html $S3_PREFIXES/index.html

for RES in ${RESULTS[*]}; do 
  for OBJ in ${files[*]}; do
    obj_path=${RES}/${OBJ}
    echo "Uploading [${result_path}/$obj_path] to [$S3_PREFIXES/$obj_path]";
    aws s3 cp "${result_path}"/"${obj_path}" "${S3_PREFIXES}"/"${obj_path}"
  done
done

