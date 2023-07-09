import 'widget.dart';
import 'dart:html' as html;

var positionX = 0;
var positionY = 0;

class Screen {
  late Widget root;

  double get width => html.window.innerWidth! as double;

  double get height => html.window.innerHeight! as double;

  void addElement(html.Element elem) {
    html.document.body!.children.add(elem);
  }

  void setRoot(Widget widget) {
    root = widget;
    // widget.init(this);
  }

  void run() {
    html.window.addEventListener("resize", (event) => render());
    html.window.addEventListener("keydown", (event) {
      var evt = event as html.KeyboardEvent;
      print(evt.key);

      if (evt.key == "ArrowRight") {
        positionX++;
        render();
      } else if (evt.key == "ArrowLeft") {
        positionX--;
        render();
      } else if (evt.key == "ArrowUp") {
        positionY--;
        render();
      } else if (evt.key == "ArrowDown") {
        positionY++;
        render();
      }
    });

    root.init(this);

    render();
  }

  void render() {
    root.update();
    // root.
  }
}
