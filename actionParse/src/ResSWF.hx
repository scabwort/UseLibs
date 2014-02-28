package ;

import format.abc.Context;
import format.abc.Writer;
import format.abc.Data;
import format.swf.Data;
import haxe.io.Bytes;
import haxe.io.BytesOutput;
import neko.zip.Compress;
import sys.io.File;
/**
 * ...
 * @author Scott Lee
 */

enum AttrVal {
	IntVal(val:Int);
	StringVal(val:String);
}
typedef AttrType = {
	var key:String;
	var val:AttrVal;
}
 
class ResSWF {
	private var swfHead:format.swf.Data.SWFHeader;
	
	private var abcCode:Context;
	private var intType:Index<Name>;
	private var stringType:Index<Name>;
	
	private var symbols : Array<SymData>;
	private var tags:Array<SWFTag>;
	private var curSymbolId:Int;
	private var clsDefs:Array<ClassDef>;
	
	private var _abcinited:Bool;
	
	public function new() {
		swfHead = { version : 10, compressed : false, width : 1, height : 1, fps : 30, nframes : 1 };
	}
	
	//建立准备工作
	public function create():Void {
		_abcinited = false;
		symbols = new Array<SymData>();
		tags = new Array<SWFTag>();
		abcCode = new Context();
		
		curSymbolId = 1;
		intType = abcCode.type("int");
		stringType = abcCode.type("String");
		
		tags.push(SWFTag.TSandBox(false, false, true, true, false));
		tags.push(SWFTag.TBackgroundColor(0x000000));
	}
	
	public function getBytes():Bytes {
		//initAbc();
		//abc 完成
		abcCode.finalize();
		var o = new BytesOutput();
		new format.abc.Writer(o).write(abcCode.getData());
		
		tags.push(SWFTag.TActionScript3(o.getBytes()));
		if (symbols.length > 0) {
			tags.push(SWFTag.TSymbolClass(symbols));
		}
		tags.push(SWFTag.TShowFrame);
		
		var outBuf = new BytesOutput();
		new format.swf.Writer(outBuf).write( { header: swfHead, tags:tags } );
		return outBuf.getBytes();
	}
	
	public function addImage(path:String, packageKey:String, needCut:Bool = true):Void {
		if (needCut) {
			Sys.command("image_trans.exe -in=" + path, []);
			var rgb = File.getBytes("tmp.jpg");
			var aData = Compress.run(File.getBytes("tmp.alpha"), 9);
			tags.push(SWFTag.TBitsJPEG(curSymbolId, JPEGData.JDJPEG3(rgb, aData)));
			symbols.push( { cid:curSymbolId,  className: packageKey});
			curSymbolId++;
			
			addImageCls(File.getContent("tmp.pos").split(","), packageKey);
		}
	}
	
	private function addImageCls(args:Array<String>, packageKey:String):Void {
		var attrs = [ { key:"x", val:IntVal(Std.parseInt(args[0])) }, { key:"y", val:IntVal(Std.parseInt(args[1])) } ];
		addClass(packageKey, attrs, "flash.display.BitmapData", [IntVal(Std.parseInt(args[2])), IntVal(Std.parseInt(args[3]))]);
	}
	
	public function addClass(packageKey:String, vals:Array<AttrType>, ?supperCls:String = null, ?defaultArgs:Array<AttrVal> = null):Void {
		var c = abcCode.beginClass(packageKey);
		if (supperCls != null) {
			c.superclass = abcCode.type(supperCls);
		}
		var ops = new Array<OpCode>();
		//构造参数
		var method:Function;
		if (defaultArgs != null) {
			method = abcCode.beginConstructor(getDefaultParametersTypes(defaultArgs));
			addDefaultParameters(defaultArgs);
			
			ops.push(OpCode.OThis);
			ops.push(OpCode.OScope);
			ops.push(OpCode.OThis);
			var argLen = defaultArgs.length;
			for (i in 0...argLen) 
			{
				ops.push(OpCode.OReg(i + 1));
			}
			ops.push(OpCode.OConstructSuper(argLen));
		} else {
			method = abcCode.beginConstructor([]);
		}
		
		method.maxStack = 3;
		method.maxScope = 1;
		
		for (attr in vals) 
		{
			switch (attr.val) 
			{
				case AttrVal.IntVal(val):
					abcCode.defineField(attr.key, intType);
					ops.push(OpCode.OThis);
					ops.push(OpCode.OInt(val));
					ops.push(OpCode.OSetProp(abcCode.property(attr.key)));
				case AttrVal.StringVal(val):
					abcCode.defineField(attr.key, stringType);
					ops.push(OpCode.OThis);
					ops.push(OpCode.OString(abcCode.string(val)));
					ops.push(OpCode.OSetProp(abcCode.property(attr.key)));
			}
		}
		ops.push(ORetVoid);
		abcCode.ops(ops);
		abcCode.endClass();
	}
	
	private function getDefaultParametersTypes(defaultArgs:Array<AttrVal>):Array<Null<IName>> {
		var parms = new Array<Null<IName>>();
		for (arg in defaultArgs) 
		{
			switch (arg) 
			{
				case IntVal(val):
					parms.push(intType);
				case StringVal(val):
					parms.push(stringType);
			}
		}
		return parms;
	}
	
	private function addDefaultParameters(defaultArgs:Array<AttrVal>):Void
	{
		var parms = new Array<Value>();
		for (arg in defaultArgs) 
		{
			switch (arg) 
			{
				case IntVal(val):
					parms.push(Value.VInt(abcCode.int(val)));
				case StringVal(val):
					parms.push(Value.VString(abcCode.string(val)));
			}
		}
		
		var curTypes = abcCode.getData().methodTypes;
		var curType = curTypes[curTypes.length -1];
		curType.extra = {
			native: false,
			variableArgs : false,
			argumentsDefined : false,
			usesDXNS : false,
			newBlock : false,
			unused : false,
			debugName : null,
			defaultParameters : parms,
			paramNames : null
		};
	}
}