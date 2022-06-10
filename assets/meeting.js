
const eyeson = window.eyeson.default;

const log = document.querySelector("#meeting-log");
const video = document.querySelector("video");
const recordingButton = document.querySelector("#record");
const videoButton = document.querySelector("#mute-video");
const audioButton = document.querySelector("#mute-audio");

// Handle events forwarded by eyeson.
eyeson.onEvent(event => {
  switch (event.type) {
    case "accept":
      video.srcObject = event.remoteStream;
      video.play();
      break;
    case "add_user":
    case "remove_user":
      const item = document.createElement("li");
      const time = (new Date()).toLocaleTimeString();
      if (event.type === "add_user") {
        item.textContent = "[" + time + "] " +
          event.user.name + " joined the meeting";
      } else {
        item.textContent = "[" + time + "] " +
          event.user.name + " left the meeting";
      }
      log.appendChild(item);
      break;
    case "recording_update":
      if (event.recording.duration) {
        recordingButton.classList.remove("active");
      } else {
        recordingButton.classList.add("active");
      }
      break;
    default:
      console.debug("received unhandled event", event.type, event);
  }
});
eyeson.start(video.dataset.accessKey);

// Handle video/audio mute buttons click.
const handleMute = event => {
  event.target.classList.toggle("muted");
  eyeson.send({
    type: "change_stream",
    video: !videoButton.classList.contains("muted"),
    audio: !audioButton.classList.contains("muted"),
  });
};
videoButton.addEventListener("click", handleMute);
audioButton.addEventListener("click", handleMute);

// Handle recording button click.
recordingButton.addEventListener("click", event => {
  const activeRecording = recordingButton.classList.contains("active");
  eyeson.send({
    type: activeRecording ? 'stop_recording' : 'start_recording',
  })
});
