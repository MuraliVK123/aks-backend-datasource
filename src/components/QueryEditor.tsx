import React, { ChangeEvent, PureComponent} from 'react';

import defaults from 'lodash/defaults';



import { Select, Switch } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from 'datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from 'types';
import _, {__} from 'lodash'
import type * as CSS from 'csstype';

interface Style extends CSS.Properties, CSS.PropertiesHyphen {}


   const options = [
    "Log","Live"
   ]

   let signalData: any[] = []
   
  
   const Options = options.map((option) => (
     <option key={option} value={option}>
       {option}
     </option>
   ));

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;


export class QueryEditor extends PureComponent<Props> {
  
  BaseURL: any;
  serverURL: string;
  variablePattern: string;
  dataType = "Log"
  editMode = false;
  editTootTip = "Enable Edit Mode"
  //onPatternChange : (event : any) => void;

  state = {selectedSignals : [],editMode : false}



  constructor(instanceSettings: Props) {
    super(instanceSettings);
    this.serverURL = instanceSettings.datasource.serverURL || '';
    this.variablePattern = instanceSettings.datasource.variablePattern || ""
    this.BaseURL = this.serverURL + "/api/realtime/"
    const query = defaults(this.props.query, defaultQuery);
    console.log("signals:" + signalData)
    query.type = query.type ? query.type : "Log"
    this.getSignals(query.type);
  }



  async getSignals(type: string){
    signalData = [];
    let logData: any[] = [];
    await fetch(this.BaseURL + type + "/signals?pattern=*")
    .then(response => response.json())
    .then(data => {
      signalData = [];
      logData = data
      let slicesData: any = logData;
     
      signalData = _.map(slicesData, function(v,i){
        return { label: v, value: v }
      })

      this.setState({selectedSignals : signalData})
    })
    .catch(error => {
      console.error(error)
    });
    
  }

  onServerChange = (event: any) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, target: event ? event.value ? event.value : event.target.value : ""});
    onRunQuery();
  };
  onDisplayNameChange = (event: ChangeEvent<HTMLInputElement>) => {

    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query,checked:event.target.checked });
    onRunQuery();
  };


  
  onDataTypeChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, type: event.target.value,target : "" });
    // executes the query
    query.target = "";
    this.getSignals(event.target.value);
    //onRunQuery();
  };


  onAliasnameChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, alias: event.target.value });
    // executes the query
    onRunQuery();

  };
  onScaleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, scale: event.target.value });
    // executes the query
    onRunQuery();


  };
  onPatternChange = (event: ChangeEvent<HTMLInputElement>) => {

    const { onChange, query, onRunQuery } = this.props;
    //query.pattern = undefined;
    //let logData: any[] = [];
    //let pattern = ""
    //let type = query.type === "Log" ? "Log":"Live";
    onChange({ ...query, pattern: event.target.value });
    onRunQuery();
  };

  onTextToggleChange = (event: any) => {
    if(this.state.editMode === false){
      this.editTootTip = "Disable Edit Mode"
    }else{
      this.editTootTip = "Enable Edit Mode"
    }
    this.setState({editMode : !this.state.editMode})
    console.log(event);

  };

  
  render() {

    const query = defaults(this.props.query, defaultQuery);
    console.log("signals:" + signalData)
    query.type = query.type ? query.type : "Log"
    //let newPattern = getTemplateSrv().replace(query.pattern, this.props.data?.request?.scopedVars);

    //console.log("signals:" + this.variablePattern,newPattern)
    
    const textAreaStyle:  Style = {
      background:'#111217',
      width: '40%'
      
    };
    const Optionsstyle:  Style = {
      width: '8%'
    };
    const patternStyle:  Style = {
      borderStyle : 'none',
      width: '40%'
    };

    const displayNameStyle:  Style = {
       top: '7px'
    };
    const aliasStyle:  Style = {
      width:'12%',
      borderStyle : 'none',
    };
    const scaleStyle:  Style = {
      width:'12%',
      borderStyle : 'none',
    };

    

    const { target,type,checked,alias,scale,pattern } = query;
    
        return (
          <div className="gf-form-group">
            <div className="gf-form">
            <label className="gf-form-label query-keyword width-10">Select Log or Live</label>
            <select  className="gf-form-input"
                 value={type} 
                 onChange={this.onDataTypeChange}
                 style = {Optionsstyle} >
             {Options}
             </select>
             {
              this.state.editMode && <textarea className="gf-form-input;width:60%" style={textAreaStyle} value={target} onChange={this.onServerChange}></textarea>
             }
             <label className="gf-form-label query-keyword width-10">Select Signal<i className="fa fa-edit" title={this.editTootTip} onClick={this.onTextToggleChange}/></label>
             {!this.state.editMode &&   
               <Select
                  className="gf-form-input"
                  isMulti={false}
                  isClearable={true}
                  width = "auto"
                  backspaceRemovesValue={false}
                  //onInputChange ={this.getsearchedSignals()}
                  onChange={this.onServerChange}
                  options={this.state.selectedSignals}
                  isSearchable={true}
                  placeholder= ""
                  value={target}
                  noOptionsMessage={'No signals found'}></Select>
                  }
                 
            </div>
            <div className="gf-form">
              <label className="gf-form-label query-keyword width-10">Display Names</label>
               <div className="css-1so7yeh width-4" style={displayNameStyle}>
               <Switch value = {checked} onChange={this.onDisplayNameChange} /> 
               </div>
            </div>
            <div className="gf-form">
             <label className="gf-form-label query-keyword width-10">Alias</label>
             <input className="gf-form-input" value={alias} onChange={this.onAliasnameChange} placeholder="Alias" style={aliasStyle} ></input>
            </div>
            <div className="gf-form">
              <label className="gf-form-label query-keyword width-10">Scale</label>
              <input className="gf-form-input" value={scale}  placeholder="Multiplier" style={scaleStyle} onChange={this.onScaleChange}></input>
              <label className="gf-form-label query-keyword width-10">Pattern</label>
              <input className="gf-form-input" value={pattern} style={patternStyle} placeholder="Redis Pattern" onChange={this.onPatternChange}></input>
            </div>
          </div>
        );
      
      
    }
  }

  

