'use strict';
var slidown = {}

slidown.init = function() {
  $('#download').on('click', slidown.onDownloadClicked);
}

slidown.onDownloadClicked = function() {
  var url = $('#url').val();
  switch (true) {
  case /(http|https):\/\/www\.slideshare\.net\/.+/.test(url):
    slidown.download(url, '/api/slideshare/download');
    break;
  case /(http|https):\/\/speakerdeck\.com\/.+/.test(url):
    slidown.download(url, '/api/speakerdeck/download');
    break;
  default:
    Materialize.toast('SlideShareかSpeakerDeckのURLを入力してください！！', 5000);
    break;
  }
}

slidown.download = function (url, api) {
  $('.message-area').hide();
  $('.searching').show();
  $('#download').addClass('disabled');
  var xhr = new XMLHttpRequest();
  xhr.open('POST', api, true);
  xhr.responseType = 'arraybuffer';
  xhr.setRequestHeader('Content-Type', 'application/json');
  xhr.addEventListener('progress', slidown.updateProgress, false);
  xhr.onreadystatechange = slidown.onReadyStateChanged;
  xhr.send(JSON.stringify({'url': url}));
}

slidown.updateProgress = function (evt) {
  if (evt.lengthComputable) {
    var percentage = evt.loaded / evt.total * 100;
    $('.determinate').css('width', percentage + '%');
  }
}

slidown.onReadyStateChanged = function () {
  switch (this.readyState) {
  case 2: // received response header.
    $('.searching').hide();
    $('#downloading-message').text('Downloading ' + slidown.bytes2str(this.getResponseHeader('Content-Length')) + '...');
    $('.downloading').show();
    break;
  case 4:
    $('.downloading').hide();
    $('#download').removeClass('disabled');
    if (this.status === 200) {
      slidown.onDownloadSuccess(this.response, this.getResponseHeader('X-FileName'));
    } else if (this.status === 201) {
      slidown.processJSON(this.response, slidown.downloadFromLink);
    } else {
      slidown.processJSON(this.response, slidown.showMessage);
    }
    break;
  }
}

slidown.onDownloadSuccess = function (ab, fileName) {
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
  slidown.showMessage('Download completed');
}

slidown.downloadFromLink = function (uri) {
  var tmpArray = uri.split('/');
  var filename = tmpArray[tmpArray.length-1];
  var link = document.createElement('a');
  link.download = filename;
  link.href = uri;
  link.click();
  slidown.showMessage('Download completed');
}

slidown.processJSON = function (ab, fn) {
  if (window.TextDecoder) {
    fn(slidown.buildErrorMsg(ab2str(ab)));
  } else {
    slidown.ab2strForIE(ab, function (str) {
      fn(slidown.buildErrorMsg(str));
    });
  }
}

slidown.buildErrorMsg = function (str) {
  try {
    return JSON.parse(str).message;
  } catch(e) {
    // network error etc...
    return '予期せぬエラーが発生しました。';
  }
}

slidown.showMessage = function (msg) {
  $('#result-message').text(msg);
  $('.message-area').show();
}

slidown.bytes2str = function (bytes) {
  var baseSize = 1024;
  if (bytes < baseSize) {
    return bytes + ' bytes';
  } else if (bytes < Math.pow(baseSize, 2)) {
    return (bytes / baseSize).toFixed(2) + ' KB';
  } else {
    return (bytes / Math.pow(baseSize, 2)).toFixed(2) + ' MB';
  }
}

slidown.ab2str = function (buf) {
  var decoder = new TextDecoder('utf-8');
  return decoder.decode(new Uint8Array(buf));
}

slidown.ab2strForIE = function (buf, callback) {
  var blob = new Blob([buf], {type:'text/plain'});
  var reader = new FileReader();
  reader.onload = function (evt) {
    callback(evt.target.result);
  };
  reader.readAsText(blob, 'utf-8');
}
