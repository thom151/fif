document.addEventListener('DOMContentLoaded', function() {

  // Use buttons to toggle between views
  document.querySelector("#music-meta-form").addEventListener('submit', create_music_meta);

  document.querySelector("#music-upload-form").addEventListener('submit', upload_music);
  // By default, load the inbox
});



async function create_music_meta(event) {
  event.preventDefault();
  const title = document.getElementById("title").value;
  const description = document.getElementById("description").value;
  const payload = {
    title: title,
    description: description,
  };
  try {

    const response = await fetch("/api/create_music_meta", {
        method:"POST",
        headers: {
            "Content-Type":"application/json"
      },
      body:JSON.stringify(payload),
    });


    if (response.ok) {
        const data = await response.json()
        const musicID = data.ID;
        console.log("musicID: ", data.ID)
        showUploadComponent(musicID);
    } else {
      console.error("Failed to save music metadata")
    }

  } catch (error) {
      console.error("Error: ", error);
 
  }
}




function showUploadComponent(musicID) {
  const  uploadForm = document.getElementById("music-upload-form");
  uploadForm.dataset.musicID = musicID;
  uploadForm.style.display = "block";
}



async function upload_music(event) {
    event.preventDefault();

  const form = event.currentTarget;
  const musicID = form.dataset.musicID || form.dataset.musicId;
  const fileInput = form.querySelector("input[type=file]");

  const formData = new FormData();
  formData.append("music", fileInput.files[0]);
  try {
      const response = await fetch(`/api/upload_music/${musicID}`, {
      method:"POST",
      body: formData
      });


      if (!response.ok) {
          throw new Error("Netowkr response was not ok")
        }
        console.log("Music Uploaded Successfully")
    } catch(error) {
      console.error("error: ", error)
    }
            

}


