import 'widget.dart';
import 'dart:html' as html;

var positionX = 0;
var positionY = 0;

class Screen {
  late Widget root;

  double get width => html.window.innerWidth! as double;

  double get height => html.window.innerHeight! as double;

  bool _redrawRequested = false;

  void addElement(html.Element elem) {
    html.document.body!.children.add(elem);
  }

  void setRoot(Widget widget) {
    root = widget;
    // widget.init(this);
  }

  void run() {
    html.window.addEventListener("resize", (event) => requestRedraw());
    html.window.addEventListener("keydown", (event) {
      var evt = event as html.KeyboardEvent;
      print(evt.key);

      if (evt.key == "ArrowRight") {
        positionX++;
        requestRedraw();
      } else if (evt.key == "ArrowLeft") {
        positionX--;
        requestRedraw();
      } else if (evt.key == "ArrowUp") {
        positionY--;
        requestRedraw();
      } else if (evt.key == "ArrowDown") {
        positionY++;
        requestRedraw();
      }
    });

    root.init(this);

    requestRedraw();
  }

  void requestRedraw() {
    if (_redrawRequested) {
      return;
    }

    _redrawRequested = true;

    html.window.requestAnimationFrame((highResTime) {
      _redrawRequested = false;

      root.update();
    });
  }
}
