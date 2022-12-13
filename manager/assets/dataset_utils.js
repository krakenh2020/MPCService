function load_datasets_table() {
  datasetsTable = document.querySelector("#datasets");

  // datasetsTable.innerHTML = "";
  // row = document.createElement("tr");
  // var th = document.createElement("th");
  // th.innerHTML = "";
  // row.appendChild(th);
  // var th2 = document.createElement("th");
  // th2.innerHTML = "Dataset";
  // row.appendChild(th2);
  // var th3 = document.createElement("th");
  // th3.innerHTML = "Data entries (size)";
  // row.appendChild(th3);
  // var th4 = document.createElement("th");
  // th4.innerHTML = "Shared with";
  // row.appendChild(th4);
  // datasetsTable.appendChild(row);

  let i = 0;
  // fetch("/datasets")

  fetch("/datasets")
    .then((response) => response.json())
    .then((datasetsList) => {
      //Once we fetch the list, we iterate over it
      datasetsList.forEach((dataset) => {
        // console.log(dataset)
        // Create the table row
        row = document.createElement("tr");

        // Create the table data elements for the species and description columns
        var checkbox = document.createElement("INPUT");
        checkbox.type = "checkbox";
        checkbox.className = "datasets";
        checkbox.value = i;
        i = i + 1;
        var check = document.createElement("td");

        var name = document.createElement("td");
        name.innerHTML = dataset.name;
        var size = document.createElement("td");
        size.innerHTML = dataset.size;
        // var cols = document.createElement("td");
        // cols.innerHTML = dataset.cols;
        var shared_nodes = document.createElement("td");
        shared_nodes.innerHTML = dataset.shared_with;

        // Add the data elements to the row
        check.appendChild(checkbox);
        row.appendChild(check);
        row.appendChild(name);
        row.appendChild(size);
        // row.appendChild(cols)
        row.appendChild(shared_nodes);

        datasetsTable.appendChild(row);
      });
    });
}

async function add_datasets() {
  // todo, get info about the dataset
  var datasetLink = document.getElementById("dataset_link").value;
  var datasetName = document.getElementById("dataset_name").value;

  const dataToSend = JSON.stringify({ name: datasetName, link: datasetLink });

  const rawResponse = await fetch("/datasets", {
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    method: "post",
    body: dataToSend,
  });
  // console.log(rawResponse)

  load_datasets_table();
}

async function getDatasets() {
  let datasets = [];
  await fetch("/datasets")
    .then((response) => response.json())
    .then((nodesList) => {
      nodesList.forEach((dataset) => {
        datasets.push([
          dataset.name,
          dataset.size,
          dataset.cols,
          dataset.shared_with,
          dataset.link,
          dataset.description,
        ]);
      });
    });
  return datasets;
}

async function loadAndSplit() {
  var fileToLoad = document.getElementById("fileToLoad").files[0];
  // console.log(fileToLoad)

  var fileReader = new FileReader();
  fileReader.onload = async function (fileLoadedEvent) {
    var data = fileLoadedEvent.target.result;

    // need to load public keys of selected MPC nodes
    let nodes = await getNodes();
    var selectedNodes = getSelectedIndexes("nodes");
    if (selectedNodes.length != 3) {
      console.log("select 3 nodes");
      return;
    }

    let res = SplitCsvText(
      data,
      nodes[selectedNodes[0]][3],
      nodes[selectedNodes[1]][3],
      nodes[selectedNodes[2]][3]
    );
    // result is an array of 4 strings: share for node 0, share for node 1, share for node 2, and description of the columns
    // this should be saved to a file with 4 lines corresponding to the returned values in respected order, see data_management/framingham_tiny_enc.txt
    // console.log("split result", res)
    download(
      res[0] + "\n" + res[1] + "\n" + res[2] + "\n" + res[3] + "\n",
      fileToLoad.name.substring(0, fileToLoad.name.length - 4) +
        "_encrypted_split_data.txt"
    );
  };

  fileReader.readAsText(fileToLoad, "UTF-8");
}
