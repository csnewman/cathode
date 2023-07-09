import 'dart:html' as html;
import 'package:cathode_tv/widgets/screen.dart';
import 'widgets/widget.dart';
import 'temp.dart';

class Poster extends Widget {
  html.ImageElement img;
  html.DivElement nameBox;

  Poster()
      : img = html.ImageElement(),
        nameBox = html.DivElement() {
    img.style.borderRadius = "10px";
    img.style.position = "absolute";
    img.style.backgroundColor = "gray";

    nameBox.style.position = "absolute";

    x = 0;
    y = 0;
    width = 0;
  }

  double _x = 0;
  double _y = 0;
  double _width = 0;
  bool _sel = false;
  String _path = "";
  String _name = "";

  double get x {
    return _x;
  }

  set x(double value) {
    _x = value;
    img.style.left = "${value}px";
    nameBox.style.left = "${value}px";
  }

  double get y {
    return _y;
  }

  set y(double value) {
    _y = value;
    img.style.top = "${value}px";
    _updateBox();
  }

  double get width {
    return _width;
  }

  set width(double value) {
    _width = value;
    img.style.width = "${value}px";
    img.style.height = "${value * 1.5}px";
    nameBox.style.width = "${value}px";

    _updateBox();
  }

  void _updateBox() {
    nameBox.style.top = "${_y + (width * 1.5) + 10}px";
  }

  bool get selected {
    return _sel;
  }

  set selected(bool value) {
    _sel = value;
    img.className = _sel ? "poster-selected" : "poster";
  }

  set path(String value) {
    _path = value;
    img.src = "images/$value";
  }

  set name(String value) {
    _name = value;
    nameBox.innerText = value;
  }

  @override
  void create() {
    screen.addElement(img);
    // screen.addElement(nameBox);
  }

  @override
  void destroy() {
    img.remove();
    // nameBox.remove();
  }
}

class Sidebar extends Widget {
  html.DivElement bg, icon;

  Sidebar()
      : bg = html.DivElement(),
        icon = html.DivElement() {
    bg.style.position = "absolute";
    bg.style.zIndex = "1000";
    bg.style.left = "0px";
    bg.style.top = "0px";
    bg.style.bottom = "0px";
    bg.style.width = "75px";
    bg.style.backgroundColor = "gray";

    icon.style.position = "absolute";
    icon.style.zIndex = "1001";
    icon.style.left = "0px";
    icon.style.top = "0px";
    icon.style.height = "75px";
    icon.style.lineHeight = "75px";
    icon.style.width = "75px";

    icon.style.backgroundColor = "green";
    icon.className = "icon";
    icon.innerText = "C";
  }

  @override
  void create() {
    screen.addElement(bg);
    screen.addElement(icon);
  }

  @override
  void update() {}

  @override
  void destroy() {
    bg.remove();
    icon.remove();
  }
}

class LibraryView extends Widget {
  html.DivElement top;
  late Sidebar sidebar;
  late List<Poster> posters;

  LibraryView() : top = html.DivElement() {
    top.style.position = "absolute";
    top.style.zIndex = "1000";
    top.style.left = "0px";
    top.style.right = "0px";
    top.style.top = "0px";
    top.style.height = "75px";
    top.style.backgroundColor = "dimgray";
  }

  @override
  void create() {
    sidebar = mount(Sidebar());
    posters = List<Poster>.empty(growable: true);

    screen.addElement(top);

    update();
  }

  @override
  void update() {
    var width = screen.width - leftPad - rightBar;
    var availableHeight = screen.height - 100;
    var posterHeight = availableHeight / 2.5;
    var posterWidth = posterHeight / 1.5;

    var xCount =
        ((width + posterSpacing) / (posterWidth + posterSpacing)).floor();
    var interPad = (width - xCount * posterWidth) / (xCount - 1);
    var yCount = 3;

    for (var i = posters.length; i < xCount * yCount; i++) {
      posters.add(Poster());
    }

    for (var i = posters.length - 1; i >= xCount * yCount; i--) {
      unmount(posters.removeLast());
    }

    if (positionX < 0) {
      positionX = 0;
    } else if (positionX >= xCount) {
      positionX = xCount - 1;
    }

    if (positionY < 0) {
      positionY = 0;
    }

    var y = 100.0;
    if (positionY > 0) {
      y += (availableHeight / 2) - (1.5 * posterHeight + interPad);
    }

    var id = 0;
    for (var j = 0; j < yCount; j++) {
      var x = leftPad;
      for (var i = 0; i < xCount; i++) {
        var poster = posters[id];

        poster.x = x;
        poster.y = y;
        poster.width = posterWidth;

        poster.selected = i == positionX &&
            ((j == 0 && positionY == 0) || (j == 1 && positionY != 0));

        var displayId = (positionY + j) * xCount + i;
        poster.path = imgs[displayId % imgs.length];
        poster.name = names[displayId % names.length];

        x += posterWidth + interPad;
        id++;
      }

      y += posterHeight + interPad;
    }

    for (var p in posters) {
      mount(p);
    }
  }

  @override
  void destroy() {}
}

const leftPad = 100.0;
const rightBar = 75.0;
const posterSpacing = 20.0;
