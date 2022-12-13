function load_nodes_table() {
  nodesTable = document.querySelector("#nodes");
  let i = 0;
  fetch("/nodes")
    .then((response) => response.json())
    .then((nodesList) => {
      //Once we fetch the list, we iterate over it
      nodesList.forEach((node) => {
        // console.log(node)
        // Create the table row
        row = document.createElement("tr");

        // Create the table data elements for the species and description columns
        var checkbox = document.createElement("INPUT");
        checkbox.type = "checkbox";
        checkbox.className = "nodes";
        checkbox.value = i;
        i = i + 1;
        var check = document.createElement("td");
        var name = document.createElement("td");
        name.innerHTML = node.name;
        var location = document.createElement("td");
        location.innerHTML = node.address;
        var description = document.createElement("td");
        description.innerHTML = node.description;

        // Add the data elements to the row
        check.appendChild(checkbox);
        row.appendChild(check);
        row.appendChild(name);
        row.appendChild(location);
        row.appendChild(description);

        nodesTable.appendChild(row);
      });
    });
}

async function getNodes() {
  let nodes = [];
  await fetch("/nodes")
    .then((response) => response.json())
    .then((nodesList) => {
      //Once we fetch the list, we iterate over it
      nodesList.forEach((node) => {
        nodes.push([
          node.name,
          node.address,
          node.scale_port,
          node.mpc_pub_key,
          node.scale_key,
        ]);
      });
    });
  return nodes;
}

function getSelectedIndexes(className) {
  var selectedNodes = [];
  var checkboxes = document.querySelectorAll("input:checked");

  for (var i = 0; i < checkboxes.length; i++) {
    if (checkboxes[i].className == className) {
      selectedNodes.push(parseInt(checkboxes[i].value));
    }
  }

  return selectedNodes;
}

function getSelectedValue(className) {
  var checkboxes = document.querySelectorAll("input:checked");

  var funcName;
  for (var i = 0; i < checkboxes.length; i++) {
    if (checkboxes[i].className == className) {
      funcName = checkboxes[i].value;
    }
  }

  return funcName;
}

// var functList = ["avg", "max", "stats"]
// demo use of MPC computation
async function mpc_computation() {
  document.getElementById("errorMsg").style.display = "none";
  // get information about selected nodes
  var selectedNodesIndexes = getSelectedIndexes("nodes");
  if (selectedNodesIndexes.length != 3) {
    document.getElementById("errorMsg").innerText = "Error: select 3 nodes.";
    document.getElementById("errorMsg").style.display = "block";
    console.log("select 3 nodes");
    return;
  }
  let allNodes = await getNodes();
  let nodes = [
    allNodes[selectedNodesIndexes[0]],
    allNodes[selectedNodesIndexes[1]],
    allNodes[selectedNodesIndexes[2]],
  ];
  var nodesNames = nodes[0][0] + "," + nodes[1][0] + "," + nodes[2][0];

  // get information about selected nodes
  var selectedDatasets = getSelectedIndexes("datasets");
  if (selectedDatasets.length == 0) {
    document.getElementById("errorMsg").innerText =
      "Error: no dataset selected.";
    document.getElementById("errorMsg").style.display = "block";
    document.getElementById("errorMsg").style.color = "red";
    console.log("no dataset selected");
    return;
  }
  let datasets = await getDatasets();
  let datasetNames = "";
  let columns = datasets[selectedDatasets[0]][2].split(",");
  let allowedNodes = nodesNames.split(",");
  for (var i = 0; i < selectedDatasets.length; i++) {
    datasetNames = datasetNames + "," + datasets[selectedDatasets[i]][0];
    columns = columns.filter((value) =>
      datasets[selectedDatasets[i]][2].split(",").includes(value)
    );
    if (datasets[selectedDatasets[i]][3] != "all") {
      allowedNodes = allowedNodes.filter((value) =>
        datasets[selectedDatasets[i]][3].split(",").includes(value)
      );
    }
  }
  datasetNames = datasetNames.substring(1);

  if (columns.length == 0) {
    document.getElementById("errorMsg").innerText =
      "Error: datasets incompatible.";
    document.getElementById("errorMsg").style.display = "block";
    document.getElementById("errorMsg").style.color = "red";
    console.log("datasets incompatible");
    return;
  }
  if (allowedNodes.length != 3) {
    document.getElementById("errorMsg").innerText =
      "Error: a dataset not shared with the selected nodes.";
    document.getElementById("errorMsg").style.display = "block";
    document.getElementById("errorMsg").style.color = "red";
    console.log("a dataset not shared with the selected nodes");
    return;
  }

  // define the name of the function that will be computed
  var funcName = getSelectedValue("function");

  var params = {};
  if (funcName == "k-means") {
    params["NUM_CLUSTERS"] = document.getElementById("num_clusters").value;
    if ((!(parseInt(params["NUM_CLUSTERS"]) > 1)) || (parseInt(params["NUM_CLUSTERS"]) > 5)) {
      document.getElementById("errorMsg").innerText =
        "Error: input of number of clusters should at least 2 and at most 5.";
      document.getElementById("errorMsg").style.display = "block";
      document.getElementById("errorMsg").style.color = "red";
      console.log("error with input of number of clusters");
      return;
    }
    document.getElementById("errorMsg").innerText =
        "Computing k-means is a complex operation that might take some time.";
    document.getElementById("errorMsg").style.display = "block";
    document.getElementById("errorMsg").style.color = "black";
  }

  var progressBar = document.querySelector("progress[id=progressBar]");
  progressBar.removeAttribute("value");

  // generate public and private key of the buyer
  let keypair = GenerateKeypair();
  let pubKey = keypair[0];
  let secKey = keypair[1];

  // send requests
  console.log("Sending requests to manager");

  var msg = {
    NodesNames: nodesNames,
    Program: funcName,
    DatasetNames: datasetNames,
    ReceiverPubKey: pubKey,
    Params: JSON.stringify(params),
  };

  // timeout 1h
  let rawResponse
  try {
    rawResponse = await fetchWithTimeout("/compute", msg, {
      timeout: 60 * 60 * 1000,
    });
  }
  catch (err) {
    document.getElementById("errorMsg").innerText =
        "Error: " + err.message;
    document.getElementById("errorMsg").style.display = "block";
    document.getElementById("errorMsg").style.color = "red";
    console.log("error computing the function");
    return;
  }


  let response = await rawResponse.json();
  console.log("Response obtained");

  let res = JoinSharesShamir(
    pubKey,
    secKey,
    response[0].Result,
    response[1].Result,
    response[2].Result
  );

  // interpret the result
  let csvText = VecToCsvText(res, response[0].Cols, funcName);
  // console.log("result", csvText)

  download(csvText, "result.csv");

  progressBar.value = 100;
  document.getElementById("errorMsg").innerText =
    "Success: see downloaded file.";
  document.getElementById("errorMsg").style.display = "block";
  document.getElementById("errorMsg").style.color = "green";
}

function download(textToWrite, name) {
  var a = document.body.appendChild(document.createElement("a"));
  a.download = name;
  textToWrite = textToWrite.replace(/\n/g, "%0D%0A");
  a.href = "data:text/plain," + textToWrite;
  a.click();
}

async function fetchWithTimeout(resource, msg, options = {}) {
  const { timeout = 8000 } = options;

  const controller = new AbortController();
  const id = setTimeout(() => controller.abort(), timeout);
  try {
    const response = await fetch(resource, {
      ...options,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "post",
      body: JSON.stringify(msg),
    });
    clearTimeout(id);
    return response;
  }
  catch (err) {
    return err.message
  }
}
