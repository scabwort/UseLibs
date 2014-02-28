package ;

import haxe.io.Bytes;
import haxe.io.BytesOutput;
import neko.Lib;
import format.swf.Data;
import format.abc.Data;
import format.abc.Context;
import neko.zip.Compress;
import nme.display.BitmapData;
import nme.geom.Rectangle;
import sys.FileSystem;
import sys.io.File;

/**
 * ...
 * @author Scott Lee
 */

 typedef ActionClass = {
	var key:String;
	var num:Int;
 }

class Main 
{
	private var swfHead:format.swf.Data.SWFHeader;
	
	private var abcCode:Context;
	private var symbols : Array<SymData>;
	private var tags:Array<SWFTag>;
	private var curClassId:Int;
	private var intType:Index<Name>;
	private var clsDefs:Array<ClassDef>;
	
	private var version:Int;
	private var isInAnimator:Bool;
	private var centerX:Int;
	private var centerY:Int;

	private var sourcePath:String;
	private var targetPath:String;
	private var tmpPath:String;
	private var targetPackage:String;
	
	private var globalWay:Int;
	private var frameIndex:Int;
	private var offsetX:Int;
	private var offsetY:Int;
	private var standQuality:Float;
	private var otherQuality:Float;
	private var oldOffsetX:Int;
	private var oldOffsetY:Int;
	private var shadow:Bool;
	var actionClass:Array<ActionClass>;
	
	private function traceHelp():Void
	{
		Lib.println("----------------------------------------");
		Lib.println("need folder of actors");
		Lib.println("----------------------------------------");
	}
	
	public function new()
	{
		swfHead = { version : 10, compressed : false, width : 1, height : 1, fps : 30, nframes : 1 };
		
		offsetX = 0;
		offsetY = 0;
		globalWay = 5;
		shadow = true;
		standQuality = 0.9;
		otherQuality = 0.9;
		var args:Array<String> = Sys.args();
		var argsCody:Array<String> = args;
		args = new Array<String>();
		var nodeArg:Array<String>;
		for (k in 0...argsCody.length) 
		{
			nodeArg = argsCody[k].split("=");
			if (nodeArg.length == 2)
			{
				switch (nodeArg[0]) 
				{
					case "shadow":
						shadow = Std.parseInt(nodeArg[1]) == 1;
					case "way":
						globalWay = Std.parseInt(nodeArg[1]);
					case "ox":
						offsetX = Std.parseInt(nodeArg[1]);
					case "oy":
						offsetY = Std.parseInt(nodeArg[1]);
					case "sq":
						standQuality = Std.parseFloat(nodeArg[1]);
					case "oq":
						otherQuality = Std.parseFloat(nodeArg[1]);
					default:
				}
			}else
			{
				args.push(argsCody[k]);
			}
		}
		
		if (args.length < 3)
		{
			traceHelp();
			return;
		}
		
		sourcePath = args[0];
		targetPath = args[1];
		targetPackage = args[2];
		
		sourcePath = sourcePath.split("\\").join("/");
		targetPath = targetPath.split("\\").join("/");
		
		Lib.println(sourcePath);
		Lib.println(targetPath);
		Lib.println(targetPackage);
		
		tmpPath = "tmp";
		if (!FileSystem.exists(sourcePath))
		{
			traceHelp();
			return;
		}
		
		var _idex:Int = sourcePath.lastIndexOf("/") + 1;
		var actorName = sourcePath.substr(_idex, sourcePath.length - _idex);
		parseActor(sourcePath, actorName, targetPath);
	}
	
	private function parseActor(path:String, type:String,targetPath:String):Void
	{
		Lib.println(path);
		var actions = FileSystem.readDirectory(path);
		createFolder(targetPath);
		
		//建立准备工作
		abcCode = new Context();
		intType = abcCode.type("int");
		symbols = new Array<SymData>();
		tags = new Array<SWFTag>();
		curClassId = 1;
		
		actionClass = new Array<ActionClass>();
		//解析图片
		var actionIndex:Int = 0;
		for (action in actions) 
		{
			actionIndex = Std.parseInt(action);
			parseAction(path + "/" + action, "res." + targetPackage + "." + type + ".a" + actionIndex, targetPath +"/" + type + "/" + actionIndex + ".swf", standQuality, actionIndex == 19, 0, Std.parseInt(action));
		}
		//testClass(1);
		//testClass(2);
		//testClass(3);
		
		//abc 完成
		abcCode.finalize();
		var o = new BytesOutput();
		new format.abc.Writer(o).write(abcCode.getData());
		var abc = o.getBytes();
		
		tags.push(SWFTag.TSandBox(false, false, false, true, false));
		tags.push(SWFTag.TActionScript3(abc));
		tags.push(SWFTag.TSymbolClass(symbols));
		tags.push(SWFTag.TShowFrame);
		
		var oswf = File.write("../../TestSwf/bin/aa.swf");
		new format.swf.Writer(oswf).write( { header:swfHead, tags:tags } );
		oswf.close();
	}
	//解析单一动作
	private function parseAction(sourcePath:String, pack:String, targetPath:String, pquality:Float = 0.6, isDeadAction:Bool = false, offsetImag:Int = 5, action:Int = 0):Void
	{
		if (!FileSystem.exists(sourcePath) || !FileSystem.isDirectory(sourcePath)) 
		{
			Lib.println("error!------------------------" + sourcePath);
			return;
		}
		var frames = FileSystem.readDirectory(sourcePath);
		Lib.println("parse Acton " + pack);
		frameIndex = 0;
		Lib.println("frames--" + frames.length);
		
		var idx:Int = pack.lastIndexOf(".");
		var picPack:String = pack + ".i";
		
		for (i in 0...frames.length) 
		{
			if (!StringTools.endsWith(frames[i], ".png")) continue;
			parseJpg2(sourcePath + "/" + frames[i], picPack + frameIndex, 256, 256, pquality);
		}
		
		actionClass.push( { key:pack + "m", num:frameIndex } );
		//createActionClass(pack + "m", frameIndex);
	}
	
	//创建位图数据
	private function parseJpg2(path:String, packPath:String, centerX:Int, centerY:Int, pquality:Float = 0.6):Void
	{
		if (!FileSystem.exists(path)) return ;
		trace("BitmapData.load --- " + path);
		/*var sourceBitmap:BitmapData = BitmapData.load(path);
		if (sourceBitmap == null)
		{
			Lib.println("image " + path + " is error!");
			return ;
		}
		
		var rect:Rectangle = sourceBitmap.getColorBoundsRect( 0xFF000000, 0, false);
		if (rect.isEmpty()) {
			sourceBitmap.dispose();
			Lib.println("image " + path + " is empty!");
			return ;
		}
		
		var px = Std.int(rect.x - sourceBitmap.width / 2 + offsetX);
		var py = Std.int(rect.y - sourceBitmap.height / 2 + offsetY);
		var pw = Std.int(rect.width);
		var ph = Std.int(rect.height);
		
		var targetBitmap = sourceBitmap.getPixels(rect);
		
		var bit1:BitmapData = new BitmapData(Std.int(rect.width), Std.int(rect.height), true);
		targetBitmap.position = 0;
		bit1.setPixels(bit1.rect, targetBitmap);
		
		var bit2:BitmapData = new BitmapData(Std.int(rect.width), Std.int(rect.height), false);
		bit2.fillRect(bit2.rect, 0xFF000000);
		bit2.draw(bit1);
		
		var tmpData = bit2.encode(BitmapData.JPG, pquality);
		
		File.saveBytes("tt.jpg", tmpData);
		Sys.command("jpegtran.exe", ["-optimize", "tt.jpg", "tt1.jpg"]);
		
		var rgbData = File.getBytes("tt1.jpg");
		
		bit1.dispose();
		bit2.dispose();
		
		var zlibBytes = Bytes.alloc(targetBitmap.length >> 2);
		for (k in 0...(targetBitmap.length >> 2)) 
		{
			zlibBytes.set(k, targetBitmap[k << 2]);
		}
		var bufByte = Compress.run(zlibBytes, 9);
		
		sourceBitmap.dispose();*/
		
		Sys.command("image_trans.exe -in=" + path, []);
		
		var rgbData = File.getBytes("tmp.jpg");
		var bufByte = File.getBytes("tmp.alpha");
		var alphaByte = Compress.run(bufByte, 9);
		
		tags.push(SWFTag.TBitsJPEG(curClassId, JPEGData.JDJPEG3(rgbData, alphaByte)));
		
		var c = abcCode.beginClass(packPath);
		c.superclass = abcCode.type("flash.display.BitmapData");
		var method:Function = abcCode.beginConstructor([intType, intType]);
		var curTypes = abcCode.getData().methodTypes;
		var curType = curTypes[curTypes.length -1];
		curType.extra = getExtra(1, 1);
		
		method.maxStack = 3;
		method.maxScope = 1;
		
		abcCode.defineField("x", intType);
		abcCode.defineField("y", intType);
		
		abcCode.ops([
			OThis,
			OScope,
			OThis,
			OReg(1),
			OReg(2),
			OConstructSuper(2),
			//OThis,
			//OInt(px),
			//OSetProp(abcCode.property("x")),
			//OThis,
			//OInt(py),
			//OSetProp(abcCode.property("y")),
			ORetVoid
		]);
		abcCode.endClass();
		
		trace("curClassId:" + curClassId + ", packPath:" + packPath);
		symbols.push({cid:curClassId,  className:packPath} );
		curClassId++;
		frameIndex++;
	}
	
	private inline function getExtra(pw:Int, ph:Int):MethodTypeExtra
	{
		return {
			native: false,
			variableArgs : false,
			argumentsDefined : false,
			usesDXNS : false,
			newBlock : false,
			unused : false,
			debugName : null,
			defaultParameters : [Value.VInt(abcCode.int(pw)),Value.VInt(abcCode.int(ph))],
			paramNames : null
		};
	}
	//Test
	private function testClass(idx:Int):Void
	{
		var c = abcCode.beginClass("res.body1.a" + idx);
		c.superclass = abcCode.type("flash.display.BitmapData");
		var method = abcCode.beginConstructor([intType, intType]);
		var curTypes = abcCode.getData().methodTypes;
		var curType = curTypes[curTypes.length -1];
		curType.extra = getExtra(100, 100);
		method.maxStack = 3;
		method.maxScope = 1;
		
		abcCode.defineField("x", intType);
		abcCode.defineField("y", intType);
		
		//abcCode.ops([
			//OThis,
			//OInt(idx),
			//OSetProp(abcCode.property("total")),
			//OThis,
			//OInt(globalWay),
			//OSetProp(abcCode.property("way")),
			//ORetVoid
		//]);
		abcCode.ops([
			OThis,
			OScope,
			OThis,
			OReg(1),
			OReg(2),
			OConstructSuper(2),
			OThis,
			OInt(10),
			OSetProp(abcCode.property("x")),
			OThis,
			OInt(20),
			OSetProp(abcCode.property("y")),
			ORetVoid
		]);
		abcCode.endClass();
	}
	//创建位图类
	private inline function createActionClass(pkey:String, totalNode:Int):Void
	{
		trace("create action class:", pkey);
		var c = abcCode.beginClass(pkey);
		var method = abcCode.beginConstructor([]);
		
		method.maxStack = 3;
		method.maxScope = 1;
		
		abcCode.defineField("total", intType);
		abcCode.defineField("way", intType);
		
		var ops = [
			OThis,
			OInt(totalNode),
			OSetProp(abcCode.property("total")),
			OThis,
			OInt(globalWay),
			OSetProp(abcCode.property("way")),
			ORetVoid
		];
		abcCode.ops(ops);
		abcCode.endClass();
		//curClassId++;
	}
	
	public function createFolder(str:String):Void
	{
		var idx1:Int = str.lastIndexOf(".");
		if (idx1 > 0)
		{
			var idx2:Int = str.lastIndexOf("/");
			if (idx2 < idx1)
			{
				str = str.substring(0, idx2);
			}
		}
		
		if (str.charAt(str.length -1) == "/")
		{
			str = str.substring(0, str.length - 1);
		}
		
		if (!FileSystem.exists(str))
		{
			createFolder(str.substring(0, str.lastIndexOf("/")));
			FileSystem.createDirectory(str);
		}
	}
	
	static function main() 
	{
		new Main();
	}
}