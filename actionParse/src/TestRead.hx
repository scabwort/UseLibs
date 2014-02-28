package ;
import format.abc.Context;
import format.abc.Data.MethodTypeExtra;
import format.swf.Reader;
import format.swf.Data;
import format.abc.Data;
import format.swf.Writer;
import haxe.io.Bytes;
import haxe.io.BytesInput;
import haxe.io.BytesOutput;
import sys.io.File;
import neko.zip.Compress;

/**
 * ...
 * @author Scott Lee
 */
class TestRead
{
	static private var intType;
	static private var curClassId:Int;

	public function new() 
	{
		var inputFile = File.read("ab.swf");
		//var inputFile = new BytesInput(File.getBytes("cc.swf"));
		var reader:format.swf.Reader = new format.swf.Reader(inputFile);
		
		var swf = reader.read();
		
		trace(swf.header);
		
		var useTags = new Array<SWFTag>();
		var symbolsData = new Array<SymData>();
		for (tag in swf.tags) 
		{
			trace("tag:" + tagStr(tag));
			switch(tag) {
			case SWFTag.TSandBox(useDirectBlit, useGpu, hasMeta, useAs3, useNetwork):
				useTags.push(tag);
			case SWFTag.TBackgroundColor( color):
				useTags.push(tag);
			case SWFTag.TBitsJPEG(id, data):
				//useTags.push(tag);
				//var abyte = Compress.run(File.getBytes("a.alpha"), 9);
				//useTags.push(SWFTag.TBitsJPEG(1, JPEGData.JDJPEG3(File.getBytes("1.jpg"), abyte)));
				useTags.push(SWFTag.TBitsJPEG(1, JPEGData.JDJPEG3(File.getBytes("1.jpg"), File.getBytes("a.alpha"))));
			case SWFTag.TActionScript3( data, context):
				trace("context:" + context);
				//useTags.push(tag);
				//useTags.push(SWFTag.TActionScript3(data));
				useTags.push(SWFTag.TActionScript3(createAbc(symbolsData)));
			case SWFTag.TSymbolClass( symbols):
				//useTags.push(tag);
				useTags.push(SWFTag.TSymbolClass(symbolsData));
			case SWFTag.TShowFrame:
				useTags.push(tag);
			default:
			}
		}
		
		var out = File.write("cc3.swf");
		new Writer(out).write( { header:swf.header, tags:useTags } );
		out.close();
		
		trace(useTags.length);
	}
	
	static function createAbc(symbolsData:Array<SymData>):Bytes
	{
		//建立准备工作
		var abcCode = new Context();
		intType = abcCode.type("int");
		
		curClassId = 1;
		addBitmapClassData(abcCode, "res.body.a1");
		symbolsData.push({ cid:curClassId, className:"res.body.a1" } );
		//addBitmapClassData(abcCode, "res.body.a2");
		//addBitmapClassData(abcCode, "res.body.a3");
		//addBitmapClassData(abcCode, "res.body.a4");
		//addBitmapClassData(abcCode, "res.body.a5");
		
		//var newSymbols
		//abc 完成
		abcCode.finalize();
		var o = new BytesOutput();
		new format.abc.Writer(o).write(abcCode.getData());
		return o.getBytes();
	}
	
	static function addBitmapClassData(abcCode:Context, pkey:String):Void
	{
		var c = abcCode.beginClass(pkey);
		c.superclass = abcCode.type("flash.display.BitmapData");
		var method = abcCode.beginConstructor([intType, intType]);
		var curTypes = abcCode.getData().methodTypes;
		var curType = curTypes[curTypes.length -1];
		curType.extra = getExtra(abcCode, 100, 100);
		
		method.maxStack = 3;
		method.maxScope = 1;
		
		abcCode.ops([
			OThis,
			OScope,
			OThis,
			OReg(1),
			OReg(2),
			OConstructSuper(2),
			ORetVoid
		]);
		abcCode.endClass();
	}
	
	static function getExtra(abcCode:Context, pw:Int, ph:Int):MethodTypeExtra
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
	
	static function tagStr(t) {
		return switch(t) {
		case SWFTag.TShowFrame:
			"[ShowFrame]";
		case SWFTag.TShape(id, data):
			"[Shape] " + id;
		case SWFTag.TMorphShape( id, data):
			"[MorphShape]" + id;
		case SWFTag.TFont( id, data):
			"[Font]" + id;
		case SWFTag.TFontInfo( id, data):
			"[FontInfo]" + id;
		case SWFTag.TBackgroundColor( color):
			"[BackgroundColor]" + color;
		case SWFTag.TDoActions(data):
			"[DoActions]";
		case SWFTag.TClip(id, frames, tags):
			"[Clip]" + id;
		case SWFTag.TPlaceObject2( po):
			"[PlaceObject2]";
		case SWFTag.TPlaceObject3( po):
			"[PlaceObject3]";
		case SWFTag.TRemoveObject2( depth):
			"[RemoveObject2]" + depth;
		case SWFTag.TFrameLabel( label, anchor):
			"[FrameLabel]" + label;
		case SWFTag.TExport( el):
			"[Export]";
		case SWFTag.TDoInitActions( id, data):
			"[DoInitActions]" + id;
		case SWFTag.TActionScript3( data, context):
			"[ActionScript3]";
		case SWFTag.TSymbolClass( symbols):
			var str = "";
			for (symbol in symbols) 
			{
				str += "[cid:" + symbol.cid + ", className:" + symbol.className + "] ";
			}
			"[SymbolClass] -- " + str;
		case SWFTag.TExportAssets( symbols):
			"[ExportAssets]";
		case SWFTag.TSandBox(useDirectBlit, useGpu, hasMeta, useAs3, useNetwork):
			"[SandBox]useDirectBlit:" + useDirectBlit + ", useGpu:" + useGpu + ", hasMeta:" + hasMeta + ", useAs3:" + useAs3 + ", useNetwork:" + useNetwork;
		case SWFTag.TBitsLossless( data):
			"[BitsLossless]";
		case SWFTag.TBitsLossless2( data):
			"[BitsLossless2]";
		case SWFTag.TBitsJPEG(id, data):
			var str =  switch (data) {
						case JPEGData.JDJPEG1(data):
							"JDJPEG1";
						case JPEGData.JDJPEG2(data):
							"JDJPEG2";
						case JPEGData.JDJPEG3(data, mask):
							"JDJPEG3";
						}
			"[BitsJPEG] -- " + str;
		case SWFTag.TJPEGTables(data):
			"[JPEGTables]";
		case SWFTag.TBinaryData( id, data):
			"[BinaryData]" + id;
		case SWFTag.TSound( data):
			"[Sound]";
		case SWFTag.TUnknown( id, data):
			"[Unknown]" + id;
		}
	}
	
	static public function main()
	{
		new TestRead();
	}
}