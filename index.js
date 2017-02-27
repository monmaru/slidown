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
    Materialize.toast('SlideShareかSpeakerDeckのURLを入力してください！！', 3000);
    break
  }
}

function download(url, api) {
  var $loading = $('#loading');
  var xhr = new XMLHttpRequest();
  $loading.show();
  xhr.open("POST", api, true);
  xhr.responseType = "blob";
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.onload = function (oEvent) {
    try {
      if (this.status == 200) {
        var blob = xhr.response;
        var fileName = xhr.getResponseHeader("X-FileName")
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
    } finally {
      $loading.hide();
    }
  };
  xhr.send(JSON.stringify({'url': url}));
}
