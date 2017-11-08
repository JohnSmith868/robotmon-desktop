import { EventEmitter } from 'fbemitter';

const GlobalEB = new EventEmitter();

class CServiceControllerEventBus extends EventEmitter {
  constructor() {
    super();
    this.EventNewItem = 'newItem';
  }
}
const CServiceControllerEB = new CServiceControllerEventBus();

class CAppEventBus extends EventEmitter {
  constructor() {
    super();
    this.EventNewEditor = 'newEditor';
  }
}
const CAppEB = new CAppEventBus();

class CEditor extends EventEmitter {
  constructor() {
    super();
    this.EventClientChanged = 'clientChanged';
  }
}
const CEditorEB = new CEditor();

class CLogs extends EventEmitter {
  constructor() {
    super();
    this.EventNewLog = 'newLog';
    this.TagDesktop = 'Desktop';
    this.TagDebug = 'Debug';
    this.LevelError = 'levelError';
    this.LevelWarning = 'levelWarning';
    this.LevelInfo = 'levelInfo';
  }
}
const CLogsEB = new CLogs();

export {
  GlobalEB,
  CServiceControllerEB,
  CAppEB,
  CEditorEB,
  CLogsEB,
};