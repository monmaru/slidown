'use strict';

$(function() {
  $('#download').on('click', onDownloadClicked);
});

function onDownloadClicked() {
  var url = $('#url').val();
  switch (true) {
  case /(http|https):\/\/www\.slideshare\.net\/.+/.test(url):
    download(url, '/api/slideshare/download');
    break;
  case /(http|https):\/\/speakerdeck\.com\/.+/.test(url):
    download(url, '/api/speakerdeck/download');
    break;
  default:
    Materialize.toast('SlideShareかSpeakerDeckのURLを入力してください！！', 5000);
    break;
  }
}

function download(url, api) {
  $('.message-area').hide();
  $('.searching').show();
  var xhr = new XMLHttpRequest();
  xhr.open('POST', api, true);
  xhr.responseType = 'arraybuffer';
  xhr.setRequestHeader('Content-Type', 'application/json');
  xhr.addEventListener('progress', updateProgress, false);
  xhr.onreadystatechange = onReadyStateChanged;
  xhr.send(JSON.stringify({'url': url}));
}

function updateProgress(evt) {
  if (evt.lengthComputable) {
    var percentage = evt.loaded / evt.total * 100;
    $('.determinate').css('width', percentage + '%');
  }
}

function onReadyStateChanged() {
  switch (this.readyState) {
  case 2: // received response header.
    $('.searching').hide();
    $('#downloading-message').text('Downloading ' + bytes2str(this.getResponseHeader('Content-Length')) + '...');
    $('.downloading').show();
    break;
  case 4:
    $('.downloading').hide();
    if (this.status === 200) {
      onDownloadSuccess(this.response, this.getResponseHeader('X-FileName'));
    } else {
      onDownloadError(this.response);
    }
    break;
  }
}

function onDownloadSuccess(ab, fileName) {
  var blob = new Blob([ab], {type: 'application/octet-binary'});
  if (window.navigator.msSaveBlob) {
    window.navigator.msSaveBlob(blob, fileName);
  } else {
    var objectURL = window.URL.createObjectURL(blob);
    var link = document.createElement('a');
    document.body.appendChild(link);
    link.href = objectURL;
    link.download = fileName;
    link.click();
    document.body.removeChild(link);
  }
  showMessage('Download completed');
}

function onDownloadError(ab) {
  if (window.TextDecoder) {
    showMessage(buildErrorMsg(ab2str(ab)));
  } else {
    ab2strForIE(ab, function(str) {
      showMessage(buildErrorMsg(str));
    });
  }
}

function buildErrorMsg(str) {
  try {
    return JSON.parse(str).message;
  } catch(e) {
    // network error etc...
    return '予期せぬエラーが発生しました。';
  }
}

function showMessage(msg) {
  $('#result-message').text(msg);
  $('.message-area').show();
}

function bytes2str(bytes) {
  var baseSize = 1024;
  if (bytes < baseSize) {
    return bytes + ' bytes';
  } else if (bytes <= Math.pow(baseSize, 2)) {
    return (bytes / baseSize).toFixed(2) + ' KB';
  } else {
    return (bytes / Math.pow(baseSize, 2)).toFixed(2) + ' MB';
  }
}

function ab2str(buf) {
  var decoder = new TextDecoder('utf-8');
  return decoder.decode(new Uint8Array(buf));
}

function ab2strForIE(buf, callback) {
  var blob = new Blob([buf], {type:'text/plain'});
  var reader = new FileReader();
  reader.onload = function(evt){callback(evt.target.result);};
  reader.readAsText(blob, 'utf-8');
}
