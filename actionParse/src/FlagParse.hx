package ;

/**
 * ...
 * @author Scott Lee
 */
class FlagParse
{
	
	static public function intVar(key:String, dvalue:Int):Int 
	{
		var args:Array<String> = Sys.args();
		for (k in 0...args.length) 
		{
			var nodeArg = args[k].split("=");
			if (nodeArg.length == 2 && nodeArg[0] == key)
			{
				return Std.parseInt(nodeArg[1]);
			}
		}
		return dvalue;
	}
	
	static public function stringVar(key:String, dvalue:String):String 
	{
		var args:Array<String> = Sys.args();
		for (k in 0...args.length) 
		{
			var nodeArg = args[k].split("=");
			if (nodeArg.length == 2 && nodeArg[0] == key)
			{
				return nodeArg[1];
			}
		}
		return dvalue;
	}
	
}