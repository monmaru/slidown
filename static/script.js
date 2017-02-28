'use strict';

$(function() {
  $('#download').on('click', onDownloadClicked);
});

function onDownloadClicked() {
  var url = $('#url').val();
  switch (true) {
  case /(http|https):\/\/www\.slideshare\.net\/.+/.test(url):
    download(url, '/api/slideshare/download')
    break;
  case /(http|https):\/\/speakerdeck\.com\/.+/.test(url):
    download(url, '/api/speakerdeck/download')
    break;
  default:
    Materialize.toast('SlideShareかSpeakerDeckのURLを入力してください！！', 5000);
    break
  }
}

function download(url, api) {
  $('.searching').show();
  var xhr = new XMLHttpRequest();
  xhr.open("POST", api, true);
  xhr.responseType = "blob";
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.addEventListener("progress", updateProgress, false);
  xhr.onreadystatechange = onReadyStateChanged;
  xhr.send(JSON.stringify({'url': url}));
}

function updateProgress(evt) {
  if(evt.lengthComputable) {
    var percentComplete = evt.loaded / evt.total * 100;
    $('.determinate').css('width', percentComplete + '%');
  }
}

function onReadyStateChanged() {
  switch (this.readyState) {
  case 2: // received response header.
    $('.searching').hide();
    $('.downloading').show();
    break;
  case 4:
    $('.downloading').hide();
    if(this.status === 200) {
      var blob = this.response;
      var fileName = this.getResponseHeader("X-FileName")
      if (window.navigator.msSaveBlob) {
        window.navigator.msSaveBlob(blob, fileName);
      }
      else {
        var objectURL = window.URL.createObjectURL(blob);
        var link = document.createElement("a");
        document.body.appendChild(link);
        link.href = objectURL;
        link.download = fileName;
        link.click();
        document.body.removeChild(link);
      }
    } else {
      alert('ERRORS!!!');
    }
    break;
  }
}
