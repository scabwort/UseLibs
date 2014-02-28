package ;
import neko.Lib;
import sys.FileSystem;
import sys.io.File;

/**
 * ...
 * @author Scott Lee
 */
class NewParser {
	private var sourcePath:String;
	private var targetPath:String;
	private var tmpPath:String;
	private var targetPackage:String;
	
	private var globalWay:Int;
	private var frameIndex:Int;
	private var offsetX:Int;
	private var offsetY:Int;
	var swf:ResSWF;
	
	public function new() {
		globalWay     = FlagParse.intVar("way", 1);
		offsetX       = FlagParse.intVar("ox", 0);
		offsetY       = FlagParse.intVar("oy", 0);
		sourcePath    = FlagParse.stringVar("in", "");
		targetPath    = FlagParse.stringVar("out", "");
		targetPackage = FlagParse.stringVar("package", "");
		
		if (!FileSystem.exists(sourcePath) || targetPath == "" || targetPackage == "") {
			Lib.println("----------------------------------------");
			Lib.println("need folder of actors");
			Lib.println("----------------------------------------");
			return;
		}
		
		swf = new ResSWF();
		
		parseActor();
	}
	
	private function parseActor():Void {
		swf.create();
		var actorName:String = sourcePath.split("/").pop();
		var actions = FileSystem.readDirectory(sourcePath);
		for (action in actions)
		{
			trace("parse action ", sourcePath + "/" + action);
			parseAction(sourcePath + "/" + action, targetPackage + "." + actorName + ".a" + action);
		}
		File.saveBytes(targetPath + "/" + actorName + ".swf", swf.getBytes());
	}
	
	private function parseAction(path:String, packageKey:String):Void {
		var files = FileSystem.readDirectory(path);
		var curIdx = 0;
		for (filePath in files)
		{
			trace("adddImage", path + "/" + filePath, "key:", packageKey + ".i" + curIdx);
			swf.addImage(path + "/" + filePath, packageKey + ".i" + curIdx);
			curIdx++;
		}
		swf.addClass(packageKey + "m", [ { key:"total", val:IntVal(files.length) }, { key:"way", val:IntVal(1) } ]);
	}
	
	
	static public function main() {
		new NewParser();
	}
}