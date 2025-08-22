document.addEventListener('DOMContentLoaded', function() {

  // Use buttons to toggle between views
  document.querySelector("#broll-meta-form").addEventListener('submit', create_broll_meta);

  document.querySelector("#broll-upload-form").addEventListener('submit', upload_broll);
  // By default, load the inbox
});



async function create_broll_meta(event) {
  event.preventDefault();
  const title = document.getElementById("title").value;
  const description = document.getElementById("description").value;
  const payload = {
    title: title,
    description: description,
  };
  try {

    const response = await fetch("/api/create_broll_meta", {
        method:"POST",
        headers: {
            "Content-Type":"application/json"
      },
      body:JSON.stringify(payload),
    });


    if (response.ok) {
        const data = await response.json()
        const brollID = data.ID;
        console.log("brollID: ", data.ID)
        showUploadComponent(brollID);
    } else {
      console.error("Failed to save broll metadata")
    }

  } catch (error) {
      console.error("Error: ", error);
 
  }
}




function showUploadComponent(brollID) {
  const  uploadForm = document.getElementById("broll-upload-form");
  uploadForm.dataset.brollID = brollID;
  uploadForm.style.display = "block";
}



async function upload_broll(event) {
    event.preventDefault();

  const form = event.target;
  const brollID = form.dataset.brollID;
  const fileInput = form.querySelector("input[type=file]");

  const formData = new FormData();
  formData.append("broll", fileInput.files[0]);
  try {
      const response = await fetch(`/api/upload_broll/${brollID}`, {
      method:"POST",
      body: formData
      });


      if (!response.ok) {
          throw new Error("Netowkr response was not ok")
        }
        console.log("Video Uploaded Successfully")
    } catch(error) {
      console.error("error: ", error)
    }
            

}


